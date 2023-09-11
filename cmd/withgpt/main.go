package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/dyluth/votes/gpt"
	"github.com/dyluth/votes/publicwhip"
	"github.com/sirupsen/logrus"
)

func main() {

	// hardcoded example tweet at the moment..
	tweet := `Great 
	@Conservatives
	 progress on electric vehicle charge points.
	
	People won’t make the switch to electric if they can’t find a charge point.
	
	This is very good progress! 
	
	My constituency hosts Britain’s first petrol station converted to EV charging, on Fulham Road, SW6.
	`

	log := logrus.New()
	apiKey := os.Getenv("APIKEY")
	if apiKey == "" {
		log.Fatal("env var APIKEY not set")
	}

	publicwhip.SetupMPs()

	resp, err := CheckTweet(apiKey, tweet, log)
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Println("got there in the end:\n" + resp)

}

func CheckTweet(apiKey, tweet string, log *logrus.Logger) (topic string, err error) {

	resp, err := AskGPT(apiKey, tweet, log)
	if err != nil {
		return "", err
	}
	log.Info(resp)

	fit := ""
	fit, topic, err = parseResponseMessage(resp)
	if err != nil {
		return "", err
	}
	log.WithField("fit", fit).Info("fit")
	log.WithField("topic", topic).Info("topic")

	// only proceed with high fits
	if fit == "high" {
		return topic, nil
	}
	return "", errors.New("fit was too low")
}

func AskGPT(apiKey, tweet string, log *logrus.Logger) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	policies := publicwhip.GetAllPolicies()

	content := "for each message, categorise the topic of the message from the following list of options:\n" +
		strings.Join(policies, "\n") +
		`\n\n  and how closely they fit from either "high" "medium" or "low" provide the following fields in a JSON dict, where applicable: topic, fit`

	messages := []gpt.Message{
		{
			Role:    "system",
			Content: content,
		},
		{
			Role:    "user",
			Content: tweet,
		},
	}

	resp, err := gpt.OpenAIRequest(ctx, apiKey, messages, log)
	if err != nil {
		log.WithError(err).Fatal("request failed")
	}

	for _, c := range resp.Choices {
		log.WithField("message", c.Message.Content).WithField("role", c.Message.Role).WithField("index", c.Index).Info("result")
	}

	return resp.Choices[len(resp.Choices)-1].Message.Content, err

}

type topicJSON struct {
	Topic string `json:"topic"`
	Fit   string `json:"fit"`
}

func parseResponseMessage(msg string) (fit, topic string, err error) {

	//try to parse it like this format:
	// {
	// 	"topic": "Incentivise Low Carbon Electricity Generation",
	// 	"fit": "high"
	// }

	topicJ := topicJSON{}
	err = json.Unmarshal([]byte(msg), &topicJ)
	if err == nil {
		return strings.ToLower(topicJ.Fit), topicJ.Topic, nil
	}

	// sometimes it also appears like:
	// topic: Incentivise Low Carbon Electricity Generation
	// fit: high

	// so try that too
	err = nil // reset error to nil
	topicRE := regexp.MustCompile(`topic:\W*(.+)\n`)
	fitRE := regexp.MustCompile(`fit:\W*(.+)\n`)
	msg = fmt.Sprintf("%v\n", msg) // needs to have at least 1 new line char at the end

	tMatch := topicRE.FindStringSubmatch(msg)
	if len(tMatch) == 2 {
		topic = tMatch[1]
	} else {
		return "", "", errors.New("cant find topic")
	}

	fMatch := fitRE.FindStringSubmatch(msg)
	if len(fMatch) == 2 {
		fit = fMatch[1]
	} else {
		return "", "", errors.New("cant find fit")
	}

	return strings.ToLower(strings.TrimSpace(fit)), strings.TrimSpace(topic), err
}

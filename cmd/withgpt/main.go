package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dyluth/votes/gpt"
	"github.com/dyluth/votes/publicwhip"
	"github.com/sirupsen/logrus"
)

func main() {

	// hardcoded example tweet at the moment..
	tweet := `It's been a fantastic evening meeting local residents in the constituency that's been my home for nearly 15 yrs

	I am campaigning to protect our green belt from being concreted over through top-down housing targets.  I am on the side of local residents. 👇🏾
	https://shorturl.at/bBGO7`

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

const (
	//PROMPT1 = "I am going to provide you with a list of categories, and then a message.  I want you to reply with what category the message best fits within.  Please keep your responses terse.  Reply with the most likely category from my list and how closely it fits from either `high` `medium` or `low`\nThe categories are: \n"
	PROMPT2 = "I am going to provide you with a list of categories, and then a message.  I want you to reply with what category the message best fits within.  Please keep your responses terse.  Reply with the most likely categories from my list and how closely they fit from either `high` `medium` or `low`\nThe categories are: \n"
)

func CheckTweet(apiKey, tweet string, log *logrus.Logger) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	policies := publicwhip.GetAllPolicies()
	content := PROMPT2 +
		strings.Join(policies, "\n") +
		"\n\nThe Message is:" +
		tweet

	messages := []gpt.Message{
		{
			Role:    "user",
			Content: content,
		},
	}

	resp, err := gpt.OpenAIRequest(ctx, apiKey, messages, log)
	if err != nil {
		log.WithError(err).Fatal("request failed")
	}
	fmt.Printf("\n%+v\n\n", resp)

	for _, c := range resp.Choices {
		log.WithField("message", c.Message.Content).WithField("role", c.Message.Role).WithField("index", c.Index).Info("result")
	}

	return resp.Choices[len(resp.Choices)-1].Message.Content, err

}

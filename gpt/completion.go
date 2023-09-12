package gpt

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/dyluth/votes/publicwhip"
	"github.com/sirupsen/logrus"
)

type Payload struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Response struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int      `json:"created"`
	Model   string   `json:"model"`
	Usage   Usage    `json:"usage"`
	Choices []Choice `json:"choices"`
	Error   *Error   `json:"error"`
}

type Error struct {
	Message string `json:"message"` //"message": "You exceeded your current quota, please check your plan and billing details.",
	Type    string `json:"type"`    //	"type": "insufficient_quota",
	// param: 	"param": null,
	//Code string: "code": "insufficient_quota"
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Every response will include a finish_reason. The possible values for finish_reason are:
//
// stop: API returned complete model output
// length: Incomplete model output due to max_tokens parameter or token limit
// content_filter: Omitted content due to a flag from our content filters
// null: API response still in progress or incomplete
type Choice struct {
	Message      ResponseMessage `json:"message"`
	FinishReason string          `json:"finish_reason"`
	Index        int             `json:"index"`
}

type ResponseMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// eg as curl from the example:
//
//	curl https://api.openai.com/v1/chat/completions \
//	 -H "Authorization: Bearer $OPENAI_API_KEY" \
//	 -H "Content-Type: application/json" \
//	 -d '{
//	 "model": "gpt-3.5-turbo",
//	 "messages": [{"role": "user", "content": "What is the OpenAI mission?"}]
//	 }'
func OpenAIRequest(ctx context.Context, apikey string, messages []Message, log *logrus.Logger) (Response, error) {
	data := Payload{
		Model:    "gpt-3.5-turbo",
		Messages: messages,
	}
	payloadBytes, err := json.Marshal(data)
	if err != nil {
		return Response{}, err
	}
	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", body)
	if err != nil {
		return Response{}, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", apikey))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 20 * time.Second,
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: 5 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 5 * time.Second,
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return Response{}, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{}, err
	}

	var r Response
	json.Unmarshal(b, &r)
	log.WithField("body", string(b)).Info("comptetion response body")
	if r.Error != nil {
		return r, fmt.Errorf(r.Error.Message)
	}
	return r, nil
}

func GetTopicOfMessage(apiKey, tweet string, log *logrus.Logger) (topic string, err error) {

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

	messages := []Message{
		{
			Role:    "system",
			Content: content,
		},
		{
			Role:    "user",
			Content: tweet,
		},
	}

	resp, err := OpenAIRequest(ctx, apiKey, messages, log)
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
		return "", "", fmt.Errorf("cant find topic: '%v'", topic)
	}

	fMatch := fitRE.FindStringSubmatch(msg)
	if len(fMatch) == 2 {
		fit = fMatch[1]
	} else {
		return "", "", errors.New("cant find fit")
	}

	return strings.ToLower(strings.TrimSpace(fit)), strings.TrimSpace(topic), err
}

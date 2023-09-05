package gpt

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

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

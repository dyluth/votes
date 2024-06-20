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
	Model          string     `json:"model"`
	Messages       []Message  `json:"messages"`
	Functions      []Function `json:"functions,omitempty"`
	FunctionToCall string     `json:"function_call,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAPIResponse struct {
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
	Index        int             `json:"index"`
	Message      ResponseMessage `json:"message"`
	FinishReason string          `json:"finish_reason"`
}

type ResponseMessage struct {
	Role         string       `json:"role"`
	Content      string       `json:"content"`
	FunctionCall FunctionCall `json:"function_call"`
}
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
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
func OpenAIRequest(ctx context.Context, apikey string, messages []Message, functions []Function, log *logrus.Logger) (OpenAPIResponse, error) {
	data := Payload{
		Model:     "gpt-4", // gpt-4o is cheaper, but doesnt seem to  be as good at picking results
		Messages:  messages,
		Functions: functions,
	}
	payloadBytes, err := json.Marshal(data)
	if err != nil {
		return OpenAPIResponse{}, err
	}
	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", body)
	if err != nil {
		return OpenAPIResponse{}, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", apikey))
	req.Header.Set("Content-Type", "application/json")

	b, err := makeAPICall(req)
	if err != nil {
		return OpenAPIResponse{}, err
	}

	var r OpenAPIResponse
	err = json.Unmarshal(b, &r)
	if err != nil {
		return OpenAPIResponse{}, err
	}
	log.WithField("body", string(b)).Info("completion response body")
	log.Debugf("completion parsed body Object: \n%+v\n\n", r)
	if r.Error != nil {
		return r, fmt.Errorf(r.Error.Message)
	}
	return r, nil
}

var makeAPICall = makeAPICallInternal

func makeAPICallInternal(req *http.Request) ([]byte, error) {
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
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func GetTopicOfMessage(apiKey, tweet string, log *logrus.Logger) (topic string, err error) {

	resp, err := AskGPT(apiKey, tweet, log)
	if err != nil {
		return "", err
	}
	log.Info(resp)

	topic, err = parseResponseMessage(resp)
	if err != nil {
		return "", err
	}
	return topic, nil
}

func AskGPT(apiKey, tweet string, log *logrus.Logger) (string, error) {
	//populate as per: https://medium.com/discovery-at-nesta/how-to-use-gpt-4-and-openais-functions-for-text-classification-ad0957be9b25
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	// we are using GetReducedPolicies to reduce our GPT prompt size (and cost!) - just using a subset of interesting policies
	policies := publicwhip.GetReducedPolicies() //GetAllPolicies()
	content := "predict the topic of the message"

	function := Function{
		Name:        "classify",
		Description: "Predict the topic of the given text",
		Parameters: FunctionParams{
			Type: "object",
			Properties: map[string]FunctionPrediction{
				"prediction": {
					Type: "array",
					Items: FunctionItems{
						Type: "string",
						Enum: policies,
					},
					Description: "the topics",
				},
			},
			Required: []string{"prediction"},
		},
	}

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

	resp, err := OpenAIRequest(ctx, apiKey, messages, []Function{function}, log)
	if err != nil {
		log.WithError(err).Fatal("request failed")
	}

	result := ""

	for _, c := range resp.Choices {
		result = c.Message.FunctionCall.Arguments
		log.WithField("message", c.Message.Content).WithField("role", c.Message.Role).WithField("index", c.Index).WithField("Arguments", c.Message.FunctionCall.Arguments).WithField("Prediction", result).Info("result")

		return result, nil
	}
	return "", errors.New("nothing found")
}

func parseResponseMessage(msg string) (topic string, err error) {

	// as a super dumb effort first
	// go through these to see if any of these match in the string output!
	for _, policy := range publicwhip.GetReducedPolicies() {
		if strings.Contains(msg, policy) {
			return policy, nil
		}
	}

	result := make(map[string]string) // try as a simple map
	err = json.Unmarshal([]byte(msg), &result)
	if err != nil {
		result := make(map[string][]string) // try as a map to list
		err = json.Unmarshal([]byte(msg), &result)
		if err != nil {
			// now try to see if it is a sentance with quotes inside
			re := regexp.MustCompile(".*\\\"(.+)\\\".*")
			matches := re.FindStringSubmatch(msg)
			if len(matches) == 2 {
				return matches[1], nil
			}
			// as a final ditch effort
			// go through these to see if any of these match in the string output!
			for _, policy := range publicwhip.GetReducedPolicies() {
				if strings.Contains(msg, policy) {
					return policy, nil
				}
			}
			return "", fmt.Errorf("could not understand GPT output")

		}
		prediction, ok := result["prediction"]
		if ok {
			if len(prediction) > 0 {

				return prediction[0], nil
			}
			return "", errors.New("prediction returned, but empty")
		}
	}

	prediction, ok := result["prediction"]
	if ok {
		return prediction, nil
	}
	return "", errors.New("prediciton not found")
}

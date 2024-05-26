package twitter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	twitter "github.com/g8rswimmer/go-twitter/v2"
)

var (
	idToName map[string]string
	client   *twitter.Client
)

const (
	idToNameFilename = "./mpTwitterIDsToName.json"
)

func Setup() error {
	// try to load from cache
	b, err := os.ReadFile(idToNameFilename)
	if err != nil {
		return err
	}
	// decode straight into idToName
	err = json.NewDecoder(bytes.NewBuffer(b)).Decode(&idToName)
	if err != nil {
		return err
	}

	token := os.Getenv("TWITTER_BEARER_TOKEN")
	if token == "" {
		return errors.New("ERROR: must have env var: TWITTER_BEARER_TOKEN set")
	}

	client = &twitter.Client{
		Authorizer: authorize{
			Token: token,
		},
		Client: http.DefaultClient,
		Host:   "https://api.twitter.com",
	}
	return nil
}

type authorize struct {
	Token string
}

func (a authorize) Add(req *http.Request) {
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", a.Token))
}

// function to get messages for an MP in the last X time
func GetMPMessages(mpTwitterID string) (map[string]*twitter.TweetDictionary, error) {

	if os.Getenv("TESTING") != "" {
		fmt.Println("USING EXAMPLE TWEETS")
		// load the file and return it
		// try to load from stored example
		b, err := os.ReadFile("./example_twitter_response.json")
		if err != nil {
			panic("cant load example file: " + err.Error())
		}
		// decode straight into mpPolicyPositions
		directory := make(map[string]*twitter.TweetDictionary)
		err = json.NewDecoder(bytes.NewBuffer(b)).Decode(&directory)
		return directory, err
	}

	opts := twitter.TweetRecentSearchOpts{
		TweetFields: []twitter.TweetField{twitter.TweetFieldConversationID, twitter.TweetFieldText},
		//StartTime:   time.Now().Add(-1 * time.Hour),
		//MaxResults: 1,
	}

	mpTwitterID = strings.TrimPrefix(mpTwitterID, "@") // need to remove any @ from userids
	query := fmt.Sprintf("from:%v -is:retweet", mpTwitterID)

	fmt.Println("Callout to tweet recent search callout")

	tweetResponse, err := client.TweetRecentSearch(context.Background(), query, opts)
	if err != nil {
		return nil, fmt.Errorf("tweet lookup error: %w", err)
	}

	return tweetResponse.Raw.TweetDictionaries(), nil
}

//function to reply to a specific message (returned by above function)

//TODO handle DMs to us

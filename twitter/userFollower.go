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
	"time"

	twitter "github.com/g8rswimmer/go-twitter/v2"
)

var (
	TwitterIdToName  map[string]string
	client           *twitter.Client
	IdToNameFilename = "./mpTwitterIDsToName.json"
)

func Setup() error {
	// try to load from cache
	b, err := os.ReadFile(IdToNameFilename)
	if err != nil {
		return err
	}
	// decode straight into idToName
	err = json.NewDecoder(bytes.NewBuffer(b)).Decode(&TwitterIdToName)
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
func GetMPMessages(mpTwitterID string, since time.Time) (map[string]*twitter.TweetDictionary, error) {

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
		StartTime:   time.Now().Add(-1 * time.Since(since)),
	}

	mpTwitterID = strings.TrimPrefix(mpTwitterID, "@") // need to remove any @ from userids
	query := fmt.Sprintf("from:%v -is:retweet", mpTwitterID)

	tweetResponse, err := client.TweetRecentSearch(context.Background(), query, opts)
	if err != nil {
		return nil, fmt.Errorf("tweet lookup error: %w", err)
	}

	return tweetResponse.Raw.TweetDictionaries(), nil
}

//TODO handle DMs to us

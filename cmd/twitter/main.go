package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/dyluth/votes/gpt"
	"github.com/dyluth/votes/publicwhip"
	"github.com/dyluth/votes/twitter"
	"github.com/sirupsen/logrus"
)

func main() {

	log := logrus.New()

	gptApiKey := os.Getenv("GPT_API_KEY")
	if gptApiKey == "" {
		log.Fatal("env var GPT_API_KEY not set")
	}

	err := twitter.Setup()
	if err != nil {
		panic(err)
	}
	publicwhip.SetupMPs(log)

	lastCheckTime := time.Now().Add(-30 * time.Minute)

	for mpTwitterHandle, mpName := range twitter.TwitterIdToName {
		log.Info(fmt.Sprintf("\nlooking at MP: %v - %v", mpTwitterHandle, mpName))
		tweets, err := twitter.GetMPMessages(mpTwitterHandle, lastCheckTime)
		if err != nil {
			panic(err)
		}
		log.Info(fmt.Sprintf("number of tweets: %v", len(tweets)))
		for id, tweet := range tweets {
			log.Info(fmt.Sprintf("tweet: %v - %v", id, tweet.Tweet.Text))

			// filter out rubbish tweets
			text, interesting := isInterestingTweet(tweet.Tweet.Text)
			if !interesting {
				log.Info("not interesting tweet.. skipping")
				continue
			}

			topic, err := gpt.GetTopicOfMessage(gptApiKey, text, log)
			if err != nil {
				panic(err)
			}
			log.Info("topic: " + topic)
			if topic == "None" {
				log.Info("None topic.. skipping ")

			} else {
				history, err := publicwhip.GetVoteHistory(mpName, topic)
				if err != nil {
					panic(err)
				}
				log.Info("history: " + history)

				message := fmt.Sprintf("%v has previously %v: %v", mpName, history, topic)
				panic("killed before posting\n" + message)
				_, err = twitter.PostMessage(message)
				if err != nil {
					panic(err)
				}

				break // only do the first good tweet
			}

		}
		if len(tweets) > 0 {
			break // only do the first MP with tweets
		}
	}

}

func isInterestingTweet(tweet string) (modifiedTweet string, interesting bool) {
	// strip out all URLs
	re := regexp.MustCompile(`\s*http(s?)://\S+\s*`)
	tweet = re.ReplaceAllString(tweet, "") // filter out tweets that are just a URL
	if len(tweet) == 0 {
		return tweet, false
	}
	// TODO replace with a list of words to avoid
	if strings.Contains("tweet", "campaign trail") { // comments about the campaign trail are rarely interesting
		return tweet, false
	}
	if strings.Contains("tweet", "condolence") { // we dont want to get involved in any message that involves sorrow
		return tweet, false
	}
	if strings.Contains("tweet", "saddened") { // we dont want to get involved in any message that involves sorrow
		return tweet, false
	}
	if strings.Contains("tweet", "leaves behind") { // we dont want to get involved in any message that involves sorrow
		return tweet, false
	}
	if strings.Contains("tweet", "sad") { // we dont want to get involved in any message that involves sorrow
		return tweet, false
	}
	return tweet, true
}

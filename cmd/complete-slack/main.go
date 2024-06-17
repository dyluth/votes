package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/dyluth/votes/approval"
	"github.com/dyluth/votes/gpt"
	"github.com/dyluth/votes/publicwhip"
	"github.com/dyluth/votes/twitter"
	"github.com/sirupsen/logrus"
)

const (
	WaitTime = 5 * time.Minute
)

var (
	log = logrus.New()
)

func main() {

	gptApiKey := os.Getenv("GPT_API_KEY")
	if gptApiKey == "" {
		log.Fatal("env var GPT_API_KEY not set")
	}

	err := twitter.Setup()
	if err != nil {
		panic(err)
	}
	publicwhip.SetupMPs(log)

	approver, err := approval.Setup(log, approved, "approvals")
	if err != nil {
		panic(err)
	}

	lastCheckTime, err := initLastCheckTime()
	if err != nil {
		panic("cant get last check time: " + err.Error())
	}

	// this needs a restructure to avoid hitting error code 429 `twitter callout status 429 Too Many Requests`
	// plan: store all the MPs in a list, with the time last queries
	// also store where in the list we got to.
	// every 20 minutes:
	//   start going through the list in order from the last point we got to (wrapping round as needed)
	//   check each MP till either done all of them or we hit a 429

	MPList := []*MPStore{}
	for mpTwitterHandle, MpTwitterName := range twitter.TwitterIdToName {
		MPList = append(MPList, &MPStore{mpTwitterHandle, MpTwitterName, lastCheckTime})
	}

	for {
		for _, mp := range MPList {
			DoMPRateLimit(log, lastCheckTime, gptApiKey, mp.MpTwitterID, mp.MpTwitterName, approver)
			mp.LastCheckTime = time.Now()
			setLastCheckTime()
		}
		log.Infof("done all MPs.. now waiting for %v", WaitTime)
		time.Sleep(WaitTime)
	}
}

func DoMPRateLimit(log *logrus.Logger, since time.Time, gptApiKey, MpTwitterID, mpName string, approver approval.Approver) {
	for {
		err := DoMP(log, since, gptApiKey, MpTwitterID, mpName, approver)
		if err != nil {
			if strings.Contains(err.Error(), "429") {
				// we have been rate limited.. stop here for 20 minutes then retry
				log.Info("Hit Twitter Rate limit - waiting 20 minutes, then continuing")
				time.Sleep(6 * time.Minute)
			}
		} else {
			return
		}
	}
}

func CheckAllMPs(log *logrus.Logger, since time.Time, gptApiKey string, approver approval.Approver) {
	for mpTwitterHandle, mpName := range twitter.TwitterIdToName {
		DoMP(log, since, gptApiKey, mpTwitterHandle, mpName, approver)
	}
}

func DoMP(log *logrus.Logger, since time.Time, gptApiKey, mpTwitterHandle, mpName string, approver approval.Approver) error {
	log.Info(fmt.Sprintf("\nlooking at MP: %v - %v", mpTwitterHandle, mpName))
	tweets, err := twitter.GetMPMessages(mpTwitterHandle, since)
	if err != nil {
		log.Warnf("failed to GetMPMessages: %v", err.Error())
		return err
	}
	log.Info(fmt.Sprintf("number of tweets: %v", len(tweets)))
	for id, tweet := range tweets {
		log.Debugf("tweet: %v - %v", id, tweet.Tweet.Text)

		// filter out rubbish tweets
		text, interesting := isInterestingTweet(tweet.Tweet.Text)
		if !interesting {
			log.Info("not interesting tweet.. skipping")
			return nil
		}

		topic, err := gpt.GetTopicOfMessage(gptApiKey, text, log)
		if err != nil {
			log.Warnf("failed to GetTopicOfMessage: %v", err.Error())
			return nil
		}
		log.Info("topic: " + topic)
		if topic == "None" {
			log.Info("None topic.. skipping ")
			return nil
		}
		history, err := publicwhip.GetVoteHistory(mpName, topic)
		if err != nil {
			log.Warnf("failed to GetVoteHistory(%v, %v): %v", mpName, topic, err.Error())
			return nil
		}
		log.Info("history: " + history)
		if strings.Contains(history, "never voted") {
			log.Info("MP has never voted on this, so skipping")
			return nil
		}

		message := fmt.Sprintf("%v has historically %v: %v", mpName, history, topic)
		// trigger manual approval
		approver.NewApprovalRequest(text, message, tweet.Tweet.ID)

	}
	return nil
}

type MPStore struct {
	MpTwitterID   string
	MpTwitterName string
	LastCheckTime time.Time
}

func isInterestingTweet(tweet string) (modifiedTweet string, interesting bool) {
	// strip out all URLs
	re := regexp.MustCompile(`\s*http(s?)://\S+\s*`)
	tweet = re.ReplaceAllString(tweet, "") // filter out tweets that are just a URL
	if len(tweet) == 0 {
		return tweet, false
	}
	wordsToAvoid := []string{"campaign trail", "condolence", "tragic", "saddened", "leaves behind", "sad", "grieve", "thank you"}
	for _, word := range wordsToAvoid {
		if strings.Contains(strings.ToLower(tweet), word) { // we dont want to get involved in any message that involves sorrow etc
			return tweet, false
		}
	}
	return tweet, true
}

var (
	lastCheckTime *time.Time
)

const (
	lastCheckTimeFileName = "./lastCheckTime.json"
)

func initLastCheckTime() (time.Time, error) {
	// try to load from cache
	if lastCheckTime != nil {
		return *lastCheckTime, nil
	}
	// create the file if it didnt exist
	file, _ := os.OpenFile(lastCheckTimeFileName, os.O_CREATE+os.O_WRONLY, os.ModePerm)
	file.Close()

	furthestBack := time.Now().Add(-1 * time.Hour)

	b, err := os.ReadFile(lastCheckTimeFileName)
	if err != nil {
		return time.Now(), err
	}

	err = json.NewDecoder(bytes.NewBuffer(b)).Decode(&lastCheckTime)
	if err != nil {
		if err.Error() == "EOF" {
			lastCheckTime = &furthestBack
		} else {
			return time.Now(), err
		}
	}
	if lastCheckTime == nil {
		lastCheckTime = &furthestBack
	}

	// only go back so far...
	if lastCheckTime.Before(furthestBack) {
		lastCheckTime = &furthestBack
	}
	return *lastCheckTime, nil
}

func setLastCheckTime() error {
	t := time.Now()
	lastCheckTime = &t

	file, _ := os.OpenFile(lastCheckTimeFileName, os.O_CREATE+os.O_WRONLY, os.ModePerm)
	defer file.Close()
	encoder := json.NewEncoder(file)
	err := encoder.Encode(t)
	return err
}

// approved is the callback when a tweet has been approved
func approved(tweetID, responseMsg string) {

	log.WithField("ID", tweetID).WithField("msg", responseMsg).Info("WOOP WOOP! approved!")

	id, err := twitter.PostReply(tweetID, responseMsg)
	if err != nil {
		log.Warnf("failed to PostMessage(): %v", err.Error())
	}
	log.Infof("Posted tweet %v", id)
}

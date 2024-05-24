package main

import (
	"fmt"
	"os"

	"github.com/dyluth/votes/gpt"
	"github.com/dyluth/votes/publicwhip"
	"github.com/sirupsen/logrus"
)

func main() {

	// tweet := `Great
	// @Conservatives
	//  progress on electric vehicle charge points.

	// People won’t make the switch to electric if they can’t find a charge point.

	// This is very good progress!

	// My constituency hosts Britain’s first petrol station converted to EV charging, on Fulham Road, SW6.
	// `
	// topic, err := calculateTopic(tweet)
	// if err != nil {
	// 	panic(err)
	// }
	mpName := "jon ashworth"

	topic := "Incentivise Low Carbon Electricity Generation" // hardcoded for testing

	position, err := getVotingHistory(topic, mpName)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\n\n%v %v: %v\n", mpName, position, topic)

}
func calculateTopic(tweet string) (string, error) {
	// hardcoded example tweet at the moment..
	log := logrus.New()
	apiKey := os.Getenv("APIKEY")
	if apiKey == "" {
		log.Fatal("env var APIKEY not set")
	}

	return gpt.GetTopicOfMessage(apiKey, tweet, log)
}

func getVotingHistory(topic, mpName string) (string, error) {
	fmt.Println("Topic: " + topic)
	publicwhip.SetupMPs()
	return publicwhip.GetVoteHistory(mpName, topic)
}

package main

import (
	"fmt"
	"os"

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

	resp, err := gpt.GetTopicOfMessage(apiKey, tweet, log)
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Println("got there in the end:\n" + resp)

}

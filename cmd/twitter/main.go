package main

import (
	"fmt"

	"github.com/dyluth/votes/twitter"
)

func main() {
	err := twitter.Setup()
	if err != nil {
		panic(err)
	}

	tweets, err := twitter.GetMPMessages("@Keir_Starmer")
	if err != nil {
		panic(err)
	}
	for id, tweet := range tweets {
		fmt.Printf("%v - %v\n", id, tweet.Tweet.Text)
	}
}

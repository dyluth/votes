package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/dyluth/votes/gpt"
	"github.com/dyluth/votes/publicwhip"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

var (
	apiKey string
	log    *logrus.Logger
)

// run as a service that can be called
// eg
// curl localhost:4000/votes -H "Content-Type: application/json" --data '{"Message":"hey we really need to tax bankers more!", "mp": "Dominic Raab"}'
func main() {

	log = logrus.New()
	apiKey = os.Getenv("APIKEY")
	if apiKey == "" {
		log.Fatal("env var APIKEY not set")
	}

	publicwhip.SetupMPs()

	r := mux.NewRouter()
	r.HandleFunc("/votes", GetRelevantVoteHandler)
	http.Handle("/", r)

	log.Info("waiting for traffic")
	err := http.ListenAndServe(":4000", r)
	log.Fatal(err)
}

type VotesBody struct {
	Message string `json:"message"`
	MP      string `json:"mp"`
}

type VotesResponse struct {
	History string `json:"message"`
	Topic   string `json:"topic"`
	Error   string `json:"error"`
}

func GetRelevantVoteHandler(w http.ResponseWriter, r *http.Request) {

	mb := VotesBody{}
	resp := &VotesResponse{}

	defer json.NewEncoder(w).Encode(resp) // make the right response when we close

	w.Header().Set("Content-Type", "application/json")

	json.NewDecoder(r.Body).Decode(&mb)

	log.WithField("who", mb.MP).WithField("message", mb.Message).Info("new request")

	topic, err := gpt.GetTopicOfMessage(apiKey, mb.Message, log)
	if err != nil {
		resp.Error = err.Error()
		return
	} else {
		resp.Topic = topic
	}
	log.WithField("topic", topic).Info("got topic")

	// TODO now we need to look up for the specific MP and find out how THEY voted for the thing
	history, err := publicwhip.GetVoteHistory(mb.MP, resp.Topic)
	if err != nil {
		resp.Error = err.Error()
		return
	} else {
		resp.History = history
	}
	log.WithField("history", history).Info("got history")

	fmt.Printf("got there in the end:\n%v\n", w)

}

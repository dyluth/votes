package twitter

import (
	"encoding/json"
	"os/exec"
)

var (
	twitterPostMessagesScriptPath = "./twitter-post-message.py"
)

// unfortunately i did these bits in python as it was easier to get running, so we now have this travesty

// python3 ./twitter-post-message.py 1795574545480286466 "ready for this?"
func PostReply(tweetID, message string) (messageID string, err error) {
	args := []string{twitterPostMessagesScriptPath}
	if tweetID != "" {
		args = append(args, tweetID)
	}
	if message != "" {
		args = append(args, message)
	}

	out, err := exec.Command("python3", args...).Output()
	if err != nil {
		return "", err
	}
	// out should be json so parse it as such
	resp := postTwitterResponse{}
	err = json.Unmarshal(out, &resp)
	if err != nil {
		return "", err
	}
	return resp.Data.ID, nil
}

func PostMessage(message string) (messageID string, err error) {
	return PostReply("", message)
}

type postTwitterResponse struct {
	Data struct {
		EditHistoryTweetIds []string `json:"edit_history_tweet_ids"`
		ID                  string   `json:"id"`
		Text                string   `json:"text"`
	} `json:"data"`
}

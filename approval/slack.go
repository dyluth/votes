package approval

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/dyluth/islack/islack"
	"github.com/sirupsen/logrus"
)

func Setup(log *logrus.Logger, approvedCallback func(tweetID, responseMsg string), approvalChannelName string) (Approver, error) {
	appToken := os.Getenv("SLACK_APP_TOKEN")
	botToken := os.Getenv("SLACK_BOT_TOKEN")
	bot, err := islack.NewSlack(appToken, botToken, log)
	if err != nil {
		panic(err)
	}
	bot.GetBotID()
	channelID, err := bot.Channels().GetChannel(approvalChannelName)
	if err != nil {
		return Approver{}, err
	}
	approver := Approver{bot: bot, channelID: channelID, log: log, approvedCallback: approvedCallback}

	//bot.SetMessageCallback(newMessage)
	//bot.SetMentionCallback(newMention)
	bot.SetReactionCallback(approver.newReaction)
	go func() {
		err := bot.Run()
		panic("SLACK DIED: " + err.Error())
	}()

	return approver, nil
}

type Approver struct {
	bot              *islack.Bot
	channelID        string
	log              *logrus.Logger
	approvedCallback func(tweetID, responseMsg string)
}

func (a *Approver) newReaction(r islack.Reaction) {

	if r.Emoji == "white_check_mark" {
		lines := strings.Split(r.Message.Text(), "±\n")
		// convert backwards:	msg := fmt.Sprintf("%v\n%v\n\n_[%v]_", tweetID, tweet, response)
		if len(lines) != 4 {
			a.log.Warnf("uh-oh, message `%v`\n should have 4 lines, not %v", r.Message.Text(), len(lines))
			return
		}
		tweetID := lines[0]
		re := regexp.MustCompile(`\_\[(.*)\]\_`)
		matches := re.FindStringSubmatch(lines[3])
		if len(matches) != 2 {
			a.log.Warnf("uh-oh, message `%v`\n should have 2 matches, not %v", lines[3], len(matches))
		}
		response := matches[1]
		go a.approvedCallback(tweetID, response)
		r.Message.ReplyThread("posted response")
	}
}

func (a *Approver) NewApprovalRequest(tweet, response, tweetID string) error {
	a.log.Info("====REQUESTING APPROVAL VIA SLACK====")
	msg := fmt.Sprintf("%v±\n%v±\n±\n_[%v]_", tweetID, tweet, response)
	_, err := a.bot.SendMessageToChannel(msg, a.channelID)
	return err
}

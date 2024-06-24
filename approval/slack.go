package approval

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/dyluth/islack/islack"
	"github.com/dyluth/votes/archive"
	"github.com/sirupsen/logrus"
)

var (
	testMode = false
)

func Setup(log *logrus.Logger, approvedCallback func(tweetID, responseMsg string), approvalChannelName string, approvalRecord *archive.Archiver, autoApprove bool) (*Approver, error) {
	appToken := os.Getenv("SLACK_APP_TOKEN")
	botToken := os.Getenv("SLACK_BOT_TOKEN")
	bot, err := islack.NewSlack(appToken, botToken, log)
	if err != nil {
		panic(err)
	}
	bot.GetBotID()
	channelID, err := bot.Channels().GetChannel(approvalChannelName)
	if err != nil {
		return &Approver{}, err
	}
	approver := Approver{
		bot:              bot,
		channelID:        channelID,
		log:              log,
		approvedCallback: approvedCallback,
		approvalRecord:   approvalRecord,
		autoApprove:      autoApprove,
	}

	//bot.SetMessageCallback(newMessage)
	//bot.SetMentionCallback(newMention)
	bot.SetReactionCallback(approver.newReaction)
	go func() {
		err := bot.Run()
		panic("SLACK DIED: " + err.Error())
	}()

	return &approver, nil
}

type Approver struct {
	bot              *islack.Bot
	channelID        string
	log              *logrus.Logger
	approvedCallback func(tweetID, responseMsg string)
	approvalRecord   *archive.Archiver
	autoApprove      bool
}

type RecordedApproval struct {
	TweetID          string
	TweetMsg         string
	ApprovedResponse string
}

var (
	EmojiList = []string{"white_check_mark", "+1", "tada", "dart"}
)

func (a *Approver) newReaction(r islack.Reaction) {
	r.Emoji = fmt.Sprintf(":%v:", r.Emoji)

	lines := strings.Split(r.Message.Text(), "\n")
	tweetID := lines[0]
	re := regexp.MustCompile(`±([\S\s]*)±`)
	matches := re.FindStringSubmatch(r.Message.Text())

	if len(matches) != 2 {
		a.log.Warnf("uh-oh, message `%v`\n should have 2 matches, not %v", r.Message.Text(), len(matches))
	}
	tweet := matches[1]
	for i := range lines {
		if strings.HasPrefix(lines[i], r.Emoji) {
			response := strings.TrimSpace(strings.TrimPrefix(lines[i], fmt.Sprintf("%v - ", r.Emoji)))
			a.approved(tweetID, tweet, response, r.Message)
		}
	}
}

func (a *Approver) NewApprovalRequest(tweet string, responses []string, tweetID string) error {
	a.log.Info("====REQUESTING APPROVAL VIA SLACK====")
	// tweet = fmt.Sprintf("<@D073DHL1PPE> - %v", tweet)  // could add in an @ notification to get our attention better
	options := []string{}
	for i := range responses {
		if i >= len(EmojiList) {
			break
		}
		options = append(options, fmt.Sprintf(":%v: - %v", EmojiList[i], responses[i]))
	}
	if len(options) == 0 {
		return errors.New("cannot request approval for a tweet with no options to tweet")
	}
	msg := fmt.Sprintf("%v\n±%v±\n\n%v", tweetID, tweet, strings.Join(options, "\n"))
	if !testMode {
		m, err := a.bot.SendMessageToChannel(msg, a.channelID)
		if err != nil {
			return err
		}
		if a.autoApprove {
			a.approved(tweetID, tweet, options[0], m)
		}
	}

	return nil
}
func (a *Approver) approved(tweetID, tweet, response string, msg islack.Message) {
	// record this approval
	approval := RecordedApproval{
		TweetID:          tweetID,
		TweetMsg:         tweet,
		ApprovedResponse: response,
	}
	a.approvalRecord.Store(approval)
	go a.approvedCallback(tweetID, response)
	if !testMode {
		msg.ReplyThread("posted response")
	}
}

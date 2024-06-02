package twitter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPostReply(t *testing.T) {
	twitterPostMessagesScriptPath = "../twitter-post-message.py" // need to set this to jump up to the parent directory
	gotMessageID, err := PostMessage("hey this is a thing")
	require.NoError(t, err)
	require.NotEmpty(t, gotMessageID)

	replyMessageID, err := PostReply(gotMessageID, "and a reply")
	require.NoError(t, err)
	require.NotEmpty(t, replyMessageID)

}

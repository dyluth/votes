package approval

import (
	"testing"

	"github.com/dyluth/islack/islack"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestApprover_newReaction(t *testing.T) {
	testMode = true
	var tweetID, responseMsg string
	c := make(chan int)
	a := Approver{
		approvedCallback: func(tweetIDin, responseMsgin string) {
			tweetID = tweetIDin
			responseMsg = responseMsgin
			c <- 0
		},
	}
	r := islack.Reaction{
		Message: islack.ConstructMessage(nil,
			`1803744491297227006
±Yesterday @Miatsf and I joined the @AgeUKLS women's group to discuss everything from housing and apprenticeships to the Conservatives' failure on the Windrush Compensation Scheme.

@ageuk does a fantastic job in supporting older people in their communities.±

:white_check_mark: - Anneliese Dodds has historically voted strongly for the policy: More funds for social care`, "", ""),
		Emoji:  "white_check_mark",
		UserID: "fakename",
	}
	a.newReaction(r)
	<-c
	require.Equal(t, "1803744491297227006", tweetID)
	require.Equal(t, "Anneliese Dodds has historically voted strongly for the policy: More funds for social care", responseMsg)

}

func TestApprover_NewApprovalRequest(t *testing.T) {
	testMode = true
	a := Approver{
		log: logrus.New(),
	}
	responses := []string{
		"Human Rights and Equality",
		"Right to strike",
		"Stop climate change",
		"Prevent abuse of zero hours contracts",
		"Homosexuality - Equal rights",
		"Welfare benefits ought rise in line with prices",
		"Do more to help refugees inclding children",
		"Higher Benefits for Ill and Disabled",
		"More funds for social care"}

	///really this test is just to check for panics in the code...
	err := a.NewApprovalRequest("tweet", responses, "12345")
	require.NoError(t, err)
}

package publicwhip

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestScrape(t *testing.T) {
	//https://www.publicwhip.org.uk/mp.php?mpn=Paul_Holmes&mpc=Eastleigh&house=commons&dmp=6703
	//https://www.publicwhip.org.uk/mp.php?mpid=42172&dmp=6703
	summary, err := GetMPPolicyPosition("42172", "6703")
	require.Equal(t, "voted strongly against the policy", summary)
	require.NoError(t, err)

	summary, err = GetMPPolicyPosition("banana", "6703")
	require.Equal(t, "", summary)
	require.Error(t, err)
}

func TestGetAllPolicies(t *testing.T) {
	//https://www.publicwhip.org.uk/mp.php?mpn=Paul_Holmes&mpc=Eastleigh&house=commons&dmp=6703
	//https://www.publicwhip.org.uk/mp.php?mpid=42172&dmp=6703
	loadAllPolicies("42172")

}

func TestInit(t *testing.T) {
	t.Skip("skipping TestInit as it fetches a full list of MPs which times out...")
	SetupMPs(logrus.New())
	require.Equal(t, len(AllMPs), 20)

}

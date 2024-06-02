package main

import (
	"testing"

	"github.com/dyluth/votes/publicwhip"
	"github.com/dyluth/votes/twitter"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func Test_main(t *testing.T) {
	twitter.IdToNameFilename = "../." + twitter.IdToNameFilename
	main()
	publicwhip.SetupMPs(logrus.New())
	history, err := publicwhip.GetVoteHistory("Wes Streeting", "Higher Pay for Public Sector Workers")
	require.NoError(t, err)
	require.Equal(t, "voted strongly for the policy", history)
}

func Test_isNotInterestingTweet(t *testing.T) {

	tweet := `In 2010, only 67% of schools in Woking were rated good or outstanding.

	Now, it’s up to 86%.
	
	In Woking and across England, only the @Conservatives have a clear plan to deliver a more prosperous future. https://t.co/giGTxuwTy8 `

	tweet, interesting := isInterestingTweet(tweet)
	require.True(t, interesting)
	require.Equal(t, tweet, `In 2010, only 67% of schools in Woking were rated good or outstanding.

	Now, it’s up to 86%.
	
	In Woking and across England, only the @Conservatives have a clear plan to deliver a more prosperous future.`)

	tweet, interesting = isInterestingTweet(` https://t.co/5mVt7IzslB `)
	require.False(t, interesting)
	require.Equal(t, "", tweet)
}

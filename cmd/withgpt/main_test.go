package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_getVotingHistory(t *testing.T) {
	mpName := "jon ashworth"
	topic := "Incentivise Low Carbon Electricity Generation" // hardcoded for testing

	position, err := getVotingHistory(topic, mpName)
	require.NoError(t, err)
	require.Equal(t, "voted strongly for the policy", position)
}

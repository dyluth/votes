package archive

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestArchiver_store(t *testing.T) {
	a := &Archiver{Filename: "./test.json"}
	err := a.Store(a)
	require.NoError(t, err)
}

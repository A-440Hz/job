package db

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewId(t *testing.T) {
	nid := NewPublicID(UserIdPrefix)
	nid2 := NewPublicID(UserIdPrefix)
	require.NotEqual(t, nid, nid2)
	require.Equal(t, len(nid), len(nid2))
	fmt.Println(nid, nid2)
	require.EqualValues(t, UserIdPrefix, nid[:3])
}

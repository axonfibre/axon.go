package axongo_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	axongo "github.com/axonfibre/axon.go/v4"
)

func TestIdentifier_Bytes(t *testing.T) {
	foo := axongo.IdentifierFromData([]byte("foo"))
	bytes, err := foo.Bytes()
	require.NoError(t, err)
	require.Len(t, bytes, axongo.IdentifierLength)

	decoded, i, err := axongo.IdentifierFromBytes(bytes)
	require.NoError(t, err)
	require.Equal(t, i, axongo.IdentifierLength)
	require.Equal(t, decoded, foo)
}

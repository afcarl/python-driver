package normalizer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNativeToNoder(t *testing.T) {
	require := require.New(t)

	f, err := getFixture("python_example_1.json")
	require.NoError(err)

	n, err := ToNode.ToNode(f)
	require.NoError(err)
	require.NotNil(n)
}

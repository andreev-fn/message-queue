package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_getDelayDuration(t *testing.T) {
	shape := []time.Duration{
		time.Minute,
		2 * time.Minute,
		4 * time.Minute,
	}

	tests := []struct {
		Retries  int
		Expected time.Duration
	}{
		{
			Retries:  0,
			Expected: time.Minute,
		},
		{
			Retries:  1,
			Expected: 2 * time.Minute,
		},
		{
			Retries:  2,
			Expected: 4 * time.Minute,
		},
		{
			Retries:  3,
			Expected: 4 * time.Minute,
		},
	}

	for _, test := range tests {
		actual := getDelayDuration(shape, test.Retries)
		require.Equal(t, test.Expected, actual)
	}
}

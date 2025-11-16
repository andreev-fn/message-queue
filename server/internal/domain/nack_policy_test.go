package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"server/internal/utils/opt"
)

func Test_pureDecide_WithBackoff(t *testing.T) {
	const maxRetries = 3
	bConf, err := NewBackoffConfig(
		[]time.Duration{time.Minute},
		opt.Some(maxRetries),
	)
	require.NoError(t, err)

	conf, err := NewQueueConfig(opt.Some(bConf), time.Minute, false)
	require.NoError(t, err)

	t.Run("NotExhaustedWithRedelivery", func(t *testing.T) {
		action := pureDecide(maxRetries-1, conf, true)
		require.Equal(t, NackActionDelay, action.Type)
		require.Equal(t, time.Minute, action.DelayDuration)
	})

	t.Run("NotExhaustedWithoutRedelivery", func(t *testing.T) {
		action := pureDecide(maxRetries-1, conf, false)
		require.Equal(t, NackActionDrop, action.Type)
	})

	t.Run("ExhaustedWithRedelivery", func(t *testing.T) {
		action := pureDecide(maxRetries, conf, true)
		require.Equal(t, NackActionDrop, action.Type)
	})

	t.Run("ExhaustedWithoutRedelivery", func(t *testing.T) {
		action := pureDecide(maxRetries, conf, false)
		require.Equal(t, NackActionDrop, action.Type)
	})
}

func Test_pureDecide_WithoutBackoff(t *testing.T) {
	conf, err := NewQueueConfig(opt.None[*BackoffConfig](), time.Minute, false)
	require.NoError(t, err)

	t.Run("WithRedelivery", func(t *testing.T) {
		action := pureDecide(0, conf, true)
		require.Equal(t, NackActionDrop, action.Type)
	})

	t.Run("WithoutRedelivery", func(t *testing.T) {
		action := pureDecide(0, conf, true)
		require.Equal(t, NackActionDrop, action.Type)
	})
}

func Test_getDelayDuration(t *testing.T) {
	shape := []time.Duration{
		1 * time.Minute,
		2 * time.Minute,
		5 * time.Minute,
	}

	tests := []time.Duration{
		1 * time.Minute,
		2 * time.Minute,
		5 * time.Minute,
		5 * time.Minute,
		5 * time.Minute,
	}

	for retryNum, expected := range tests {
		actual := getDelayDuration(shape, retryNum)
		require.Equal(t, expected, actual)
	}
}

func Test_getDelayDuration_ConstShape(t *testing.T) {
	shape := []time.Duration{
		time.Minute,
	}

	tests := []time.Duration{
		time.Minute,
		time.Minute,
		time.Minute,
	}

	for retryNum, expected := range tests {
		actual := getDelayDuration(shape, retryNum)
		require.Equal(t, expected, actual)
	}
}

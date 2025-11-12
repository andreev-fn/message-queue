package postgres_test

import (
	"context"
	"testing"
	"time"

	"server/internal/eventbus/postgres"
	"server/internal/utils/testutils"

	"github.com/stretchr/testify/require"
)

const unwantedChannel = "other"
const testChannel = "test"
const testMessage1 = "hello world"
const testMessage2 = "abacaba"

func TestPubSubDriverTwoMessages(t *testing.T) {
	testutils.SkipIfNotIntegration(t)

	db, err := testutils.OpenDB()
	require.NoError(t, err)

	driver := postgres.NewPubSubDriver(db)

	publish1Err := make(chan error, 1)
	publish2Err := make(chan error, 1)

	go func() {
		time.Sleep(time.Millisecond * 50)
		publish1Err <- driver.Publish(testChannel, testMessage1)
		publish2Err <- driver.Publish(testChannel, testMessage2)
	}()

	var lastReceivedChannel *string
	var lastReceivedMessage *string
	var receiveCount int

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	err = driver.Listen(ctx, []string{testChannel}, func(channel string, message string) {
		receiveCount++
		lastReceivedChannel = &channel
		lastReceivedMessage = &message
	})
	require.ErrorIs(t, err, context.DeadlineExceeded)

	require.Equal(t, 2, receiveCount)
	require.NotNil(t, lastReceivedChannel)
	require.Equal(t, testChannel, *lastReceivedChannel)
	require.NotNil(t, lastReceivedMessage)
	require.Equal(t, testMessage2, *lastReceivedMessage)

	require.NoError(t, <-publish1Err)
	require.NoError(t, <-publish2Err)
}

func TestPubSubDriverUnwantedMessage(t *testing.T) {
	testutils.SkipIfNotIntegration(t)

	db, err := testutils.OpenDB()
	require.NoError(t, err)

	driver := postgres.NewPubSubDriver(db)

	ctx1, cancel1 := context.WithTimeout(context.Background(), time.Millisecond*50)
	defer cancel1()

	// subscribe to an unwanted channel first, to check that unsubscription works properly
	err = driver.Listen(ctx1, []string{unwantedChannel}, func(channel string, message string) {})
	require.ErrorIs(t, err, context.DeadlineExceeded)

	publishErr := make(chan error, 1)

	go func() {
		time.Sleep(time.Millisecond * 50)
		publishErr <- driver.Publish(unwantedChannel, testMessage1)
	}()

	var receivedChannel *string
	var receivedMessage *string
	var receiveCount int

	ctx2, cancel2 := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel2()

	err = driver.Listen(ctx2, []string{testChannel}, func(channel string, message string) {
		receiveCount++
		receivedChannel = &channel
		receivedMessage = &message
	})
	require.ErrorIs(t, err, context.DeadlineExceeded)

	require.Equal(t, 0, receiveCount)
	require.Nil(t, receivedChannel)
	require.Nil(t, receivedMessage)

	require.NoError(t, <-publishErr)
}

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/signal"
	"slices"
	"sync/atomic"
	"syscall"
	"time"
)

// batchSizePublish = 1, batchSizeConsume = 100, batchSizeAck = 1
// 2024/12/26 02:03:17 published 500038 messages (4347.6 t/s), consumed 499732 messages (4344.9 t/s)
// 2025/08/09 12:48:59 published 861166 messages (2562.8 t/s), consumed 860965 messages (2568.2 t/s)

// batchSizePublish = 1, batchSizeConsume = 10, batchSizeAck = 1
// 2025/08/23 17:35:48 published 710189 messages (2574.4 t/s), consumed 710066 messages (2576.4 t/s)

// batchSizePublish = 100, batchSizeConsume = 100, batchSizeAck = 100
// 2025/08/23 17:28:31 published 3874300 messages (12579.6 t/s), consumed 1694600 messages (5339.8 t/s)

const threadsCountW = 24
const threadsCountR = 32

const batchSizePublish = 1
const batchSizeConsume = 10
const batchSizeAck = 1

var publishedCount atomic.Int32
var consumedCount atomic.Int32

var httpClient = &http.Client{
	Transport: &http.Transport{
		MaxIdleConnsPerHost: 100,
	},
	Timeout: 10 * time.Second,
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	for i := 0; i < threadsCountW; i++ {
		go runPublisher(ctx, i)
	}
	for i := 0; i < threadsCountR; i++ {
		go runConsumer(ctx)
	}

	go printStats(ctx)

	<-ctx.Done()
}

func printStats(ctx context.Context) {
	prevPublished := int32(0)
	prevConsumed := int32(0)
	prevTime := time.Now()

	for {
		if ctx.Err() != nil {
			break
		}

		time.Sleep(time.Second * 5)

		published := publishedCount.Load()
		consumed := consumedCount.Load()
		elapsed := time.Since(prevTime)

		publishRate := float64(published-prevPublished) / elapsed.Seconds()
		consumeRate := float64(consumed-prevConsumed) / elapsed.Seconds()

		prevPublished = published
		prevConsumed = consumed
		prevTime = time.Now()

		log.Printf(
			"published %d messages (%.1f t/s), consumed %d messages (%.1f t/s)",
			published, publishRate, consumed, consumeRate,
		)
	}
}

const baseURL = "http://127.0.0.1:8060"

type Message struct {
	ID string `json:"id"`
}

func runPublisher(ctx context.Context, threadNum int) {
	if threadNum > 99 {
		panic("too many threads")
	}

	argSequence := 0

	for {
		if ctx.Err() != nil {
			break
		}

		args := make([]string, 0, batchSizePublish)

		for i := 0; i < batchSizePublish; i++ {
			args = append(args, fmt.Sprintf("%.2d%.10d", threadNum, argSequence))
			argSequence += 1
		}

		if err := publishMessages(args); err != nil {
			log.Println(err)
			continue
		}

		publishedCount.Add(int32(len(args)))
	}
}

func publishMessages(args []string) error {
	reqElements := make([]any, 0, len(args))
	for _, arg := range args {
		reqElements = append(reqElements, map[string]any{
			"queue": "test",
			"payload": map[string]any{
				"arg": arg,
			},
		})
	}
	requestBody, err := json.Marshal(reqElements)
	if err != nil {
		return err
	}

	request, err := http.NewRequest(
		http.MethodPost,
		baseURL+"/messages/publish",
		bytes.NewReader(requestBody),
	)
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")

	response, err := httpClient.Do(request)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"response return %d status code; server response: %s",
			response.StatusCode,
			string(responseBody),
		)
	}

	return nil
}

func runConsumer(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			break
		}

		messages, err := consumeMessages(batchSizeConsume)
		if err != nil {
			log.Println("consume message error:", err)
			continue
		}

		for chunk := range slices.Chunk(messages, batchSizeAck) {
			ids := make([]string, 0, len(chunk))

			for _, message := range chunk {
				ids = append(ids, message.ID)
			}

			if err := ackMessages(ids); err != nil {
				log.Println("ack messages error:", err)
				continue
			}

			consumedCount.Add(int32(len(ids)))
		}
	}
}

func consumeMessages(limit int) ([]Message, error) {
	requestBody, err := json.Marshal(map[string]any{
		"queue": "test",
		"limit": limit,
	})
	if err != nil {
		return nil, fmt.Errorf("json.Marshal: %w", err)
	}

	request, err := http.NewRequest(http.MethodPost, baseURL+"/messages/consume", bytes.NewReader(requestBody))
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/json")

	httpResponse, err := httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("httpClient.Do: %w", err)
	}

	defer httpResponse.Body.Close()

	responseBody, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, fmt.Errorf("io.ReadAll: %w", err)
	}

	if httpResponse.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-200 response: %s, body: %s", httpResponse.Status, string(responseBody))
	}

	type ResponseDTO struct {
		Success *bool     `json:"success"`
		Result  []Message `json:"result"`
	}

	var responseDTO ResponseDTO
	err = json.Unmarshal(responseBody, &responseDTO)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal: %w; response: %s", err, string(responseBody))
	}

	if responseDTO.Success == nil {
		return nil, errors.New("invalid response")
	}

	if *responseDTO.Success == false {
		return nil, errors.New("unsuccessful response")
	}

	return responseDTO.Result, nil
}

func ackMessages(ids []string) error {
	reqElements := make([]any, 0, len(ids))
	for _, id := range ids {
		reqElements = append(reqElements, map[string]any{
			"id": id,
		})
	}
	requestBody, err := json.Marshal(reqElements)
	if err != nil {
		return fmt.Errorf("json.Marshal: %w", err)
	}

	request, err := http.NewRequest(http.MethodPost, baseURL+"/messages/ack", bytes.NewReader(requestBody))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")

	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("httpClient.Do: %w", err)
	}

	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("io.ReadAll: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"response return %d status code; server response: %s",
			response.StatusCode,
			string(responseBody),
		)
	}

	return nil
}

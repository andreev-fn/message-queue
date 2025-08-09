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
	"sync/atomic"
	"syscall"
	"time"
)

// 2024/12/26 02:03:17 created 500038 tasks (4347.6 t/s), consumed 499732 tasks (4344.9 t/s)
// 2025/08/09 12:48:59 created 861166 tasks (2562.8 t/s), consumed 860965 tasks (2568.2 t/s)

const threadsCountW = 24
const threadsCountR = 32

var createdCount atomic.Int32
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
		go createTasks(ctx, i)
	}
	for i := 0; i < threadsCountR; i++ {
		go consumeTasks(ctx)
	}

	go printStats(ctx)

	<-ctx.Done()
}

func printStats(ctx context.Context) {
	prevCreated := int32(0)
	prevConsumed := int32(0)
	prevTime := time.Now()

	for {
		if ctx.Err() != nil {
			break
		}

		time.Sleep(time.Second * 5)

		created := createdCount.Load()
		consumed := consumedCount.Load()
		elapsed := time.Since(prevTime)

		createRate := float64(created-prevCreated) / elapsed.Seconds()
		consumeRate := float64(consumed-prevConsumed) / elapsed.Seconds()

		prevCreated = created
		prevConsumed = consumed
		prevTime = time.Now()

		log.Printf(
			"created %d tasks (%.1f t/s), consumed %d tasks (%.1f t/s)",
			created, createRate, consumed, consumeRate,
		)
	}
}

const baseURL = "http://127.0.0.1:8060"

type Task struct {
	ID string `json:"id"`
}

func createTasks(ctx context.Context, threadNum int) {
	if threadNum > 99 {
		panic("too many threads")
	}

	i := 0

	for {
		if ctx.Err() != nil {
			break
		}

		arg := fmt.Sprintf("%.2d%.10d", threadNum, i)
		i += 1

		if err := createTask(arg); err != nil {
			log.Println(err)
			continue
		}

		createdCount.Add(1)
	}
}

func createTask(arg string) error {
	requestBody, err := json.Marshal(map[string]any{
		"kind": "test",
		"payload": map[string]any{
			"arg": arg,
		},
		"auto_confirm": true,
	})
	if err != nil {
		return err
	}

	request, err := http.NewRequest(
		http.MethodPost,
		baseURL+"/task/create",
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

func consumeTasks(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			break
		}

		if err := consumeTask(); err != nil {
			log.Println("consume task error:", err)
			continue
		}
	}
}

func consumeTask() error {
	tasks, err := takeWork()
	if err != nil {
		return fmt.Errorf("takeWork: %w", err)
	}

	for _, task := range tasks {
		if err := finishWork(task.ID); err != nil {
			return fmt.Errorf("finishWork: %w", err)
		}

		consumedCount.Add(1)
	}

	return nil
}

func takeWork() ([]Task, error) {
	request, err := http.NewRequest(http.MethodPost, baseURL+"/work/take?kind=test&limit=100", nil)
	if err != nil {
		return nil, err
	}

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

	type GetTasksResponse struct {
		Success *bool  `json:"success"`
		Jobs    []Task `json:"result"`
	}

	var getTasksResponse GetTasksResponse
	err = json.Unmarshal(responseBody, &getTasksResponse)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal: %w; response: %s", err, string(responseBody))
	}

	if getTasksResponse.Success == nil {
		return nil, errors.New("invalid response")
	}

	if *getTasksResponse.Success == false {
		return nil, errors.New("unsuccessful response")
	}

	return getTasksResponse.Jobs, nil
}

func finishWork(id string) error {
	requestBody, err := json.Marshal(map[string]any{
		"id":     id,
		"report": "{\"result\":\"success\"}",
	})
	if err != nil {
		return fmt.Errorf("json.Marshal: %w", err)
	}

	request, err := http.NewRequest(
		http.MethodPost,
		baseURL+"/work/finish",
		bytes.NewReader(requestBody),
	)
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

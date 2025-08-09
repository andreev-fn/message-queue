package test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"server/internal/domain"
	"server/test/e2eutils"
)

func TestCreateTask(t *testing.T) {
	app, _ := e2eutils.Prepare(t)

	const (
		taskKind     = "test"
		taskPayload  = `{"arg": 123}`
		taskPriority = 100
	)

	// Act
	body, err := json.Marshal(map[string]any{
		"kind":    taskKind,
		"payload": json.RawMessage(taskPayload),
	})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, "/task/create", bytes.NewBuffer(body))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	app.Router.ServeHTTP(resp, req)

	// Assert response
	require.Equal(t, http.StatusOK, resp.Result().StatusCode)

	var respWrapper e2eutils.ResponseWrapper
	err = json.NewDecoder(resp.Body).Decode(&respWrapper)
	require.NoError(t, err)

	require.True(t, respWrapper.Success)
	require.NotNil(t, respWrapper.Result)
	require.Nil(t, respWrapper.Error)

	var respDTO struct {
		ID string `json:"id"`
	}
	err = json.Unmarshal(*respWrapper.Result, &respDTO)
	require.NoError(t, err)

	// Assert task in DB
	task, err := app.TaskRepo.GetTaskByID(context.Background(), app.DB, respDTO.ID)
	require.NoError(t, err)

	require.Equal(t, taskKind, task.Kind())
	require.JSONEq(t, taskPayload, string(task.Payload()))
	require.Equal(t, app.Clock.Now(), task.CreatedAt())
	require.Equal(t, domain.TaskStatusCreated, task.Status())
	require.Equal(t, taskPriority, task.Priority())
}

func TestCreateTaskWithPriorityAndConfirm(t *testing.T) {
	app, _ := e2eutils.Prepare(t)

	const (
		taskKind     = "test"
		taskPayload  = `{"arg": 123}`
		taskPriority = 5
	)

	// Act
	body, err := json.Marshal(map[string]any{
		"kind":         taskKind,
		"payload":      json.RawMessage(taskPayload),
		"auto_confirm": true,
		"priority":     taskPriority,
	})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, "/task/create", bytes.NewBuffer(body))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	app.Router.ServeHTTP(resp, req)

	// Assert response
	require.Equal(t, http.StatusOK, resp.Result().StatusCode)

	var respWrapper e2eutils.ResponseWrapper
	err = json.NewDecoder(resp.Body).Decode(&respWrapper)
	require.NoError(t, err)

	require.True(t, respWrapper.Success)
	require.NotNil(t, respWrapper.Result)
	require.Nil(t, respWrapper.Error)

	var respDTO struct {
		ID string `json:"id"`
	}
	err = json.Unmarshal(*respWrapper.Result, &respDTO)
	require.NoError(t, err)

	// Assert task in DB
	task, err := app.TaskRepo.GetTaskByID(context.Background(), app.DB, respDTO.ID)
	require.NoError(t, err)

	require.Equal(t, taskKind, task.Kind())
	require.JSONEq(t, taskPayload, string(task.Payload()))
	require.Equal(t, app.Clock.Now(), task.CreatedAt())
	require.Equal(t, domain.TaskStatusReady, task.Status())
	require.Equal(t, taskPriority, task.Priority())
}

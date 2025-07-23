package test

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"server/internal/domain"
	"testing"
)

func TestCreateTask(t *testing.T) {
	app, _ := buildTestApp(t)
	cleanupDatabase(t, app.DB)

	body, err := json.Marshal(map[string]any{
		"kind": "test",
		"payload": map[string]any{
			"arg": "1234",
		},
	})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, "/task/create", bytes.NewBuffer(body))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")

	// Act
	resp := httptest.NewRecorder()
	app.Router.ServeHTTP(resp, req)

	// Assert
	require.Equal(t, http.StatusOK, resp.Result().StatusCode)

	var respWrapper ResponseWrapper
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

	// Verify task in database
	task, err := app.TaskRepo.GetTaskByID(context.Background(), app.DB, respDTO.ID)
	require.NoError(t, err)

	require.Equal(t, "test", task.Kind())
	require.JSONEq(t, `{"arg": "1234"}`, string(task.Payload()))
	require.Equal(t, app.Clock.Now(), task.CreatedAt())
	require.Equal(t, domain.TaskStatusCreated, task.Status())
	require.Equal(t, 100, task.Priority())
}

func TestCreateTaskWithPriorityAndConfirm(t *testing.T) {
	app, _ := buildTestApp(t)
	cleanupDatabase(t, app.DB)

	body, err := json.Marshal(map[string]any{
		"kind": "test",
		"payload": map[string]any{
			"arg": "12345",
		},
		"auto_confirm": true,
		"priority":     5,
	})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, "/task/create", bytes.NewBuffer(body))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")

	// Act
	resp := httptest.NewRecorder()
	app.Router.ServeHTTP(resp, req)

	// Assert
	require.Equal(t, http.StatusOK, resp.Result().StatusCode)

	var respWrapper ResponseWrapper
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

	// Verify task in database
	task, err := app.TaskRepo.GetTaskByID(context.Background(), app.DB, respDTO.ID)
	require.NoError(t, err)

	require.Equal(t, "test", task.Kind())
	require.JSONEq(t, `{"arg": "12345"}`, string(task.Payload()))
	require.Equal(t, app.Clock.Now(), task.CreatedAt())
	require.Equal(t, domain.TaskStatusReady, task.Status())
	require.Equal(t, 5, task.Priority())
}

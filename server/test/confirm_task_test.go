package test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"server/internal/domain"
	"server/test/e2eutils"
)

func TestConfirmTask(t *testing.T) {
	app, _ := e2eutils.Prepare(t)

	const (
		taskKind     = "test"
		taskPayload  = `{"arg": 123}`
		taskPriority = 100
	)

	// Arrange
	taskID := e2eutils.CreateTask(t, app, taskKind, taskPayload, taskPriority)

	// Act
	req, err := http.NewRequest(http.MethodPost, "/task/confirm?id="+taskID, nil)
	require.NoError(t, err)

	resp := httptest.NewRecorder()
	app.Router.ServeHTTP(resp, req)

	// Assert response
	require.Equal(t, http.StatusOK, resp.Result().StatusCode)

	var respWrapper e2eutils.ResponseWrapper
	err = json.NewDecoder(resp.Body).Decode(&respWrapper)
	require.NoError(t, err)

	require.True(t, respWrapper.Success)
	require.Nil(t, respWrapper.Result)
	require.Nil(t, respWrapper.Error)

	// Assert task in DB
	task, err := app.TaskRepo.GetTaskByID(context.Background(), app.DB, taskID)
	require.NoError(t, err)

	require.Equal(t, taskKind, task.Kind())
	require.JSONEq(t, taskPayload, string(task.Payload()))
	require.Equal(t, taskPriority, task.Priority())
	require.Equal(t, app.Clock.Now(), task.CreatedAt())
	require.Equal(t, domain.TaskStatusReady, task.Status())
}

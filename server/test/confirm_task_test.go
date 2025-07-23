package test

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"server/internal/domain"
	"testing"
)

func TestConfirmTask(t *testing.T) {
	app, _ := buildTestApp(t)
	cleanupDatabase(t, app.DB)

	const taskID = "00000000-0000-0000-0000-000000000001"
	testTask := createTask(t, app, taskID, "test", 100)

	err := app.TaskRepo.SaveInNewTransaction(context.Background(), app.DB, testTask)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, "/task/confirm?id="+taskID, nil)
	require.NoError(t, err)

	resp := httptest.NewRecorder()
	app.Router.ServeHTTP(resp, req)

	// Assert
	require.Equal(t, http.StatusOK, resp.Result().StatusCode)

	var respWrapper ResponseWrapper
	err = json.NewDecoder(resp.Body).Decode(&respWrapper)
	require.NoError(t, err)

	require.True(t, respWrapper.Success)
	require.Nil(t, respWrapper.Result)
	require.Nil(t, respWrapper.Error)

	// Verify task in database
	task, err := app.TaskRepo.GetTaskByID(context.Background(), app.DB, taskID)
	require.NoError(t, err)

	require.Equal(t, "test", task.Kind())
	require.JSONEq(t, `{"arg": "cf404dc806178c245b5b4fe2531e6d8c"}`, string(task.Payload()))
	require.Equal(t, app.Clock.Now(), task.CreatedAt())
	require.Equal(t, domain.TaskStatusReady, task.Status())
}

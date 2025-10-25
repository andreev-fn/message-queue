package routes

import (
	"context"
	"log/slog"
	"net/http"

	"server/internal/routes/base"
	"server/internal/usecases"
	"server/pkg/httpmodels"
)

type CheckMessages struct {
	logger  *slog.Logger
	useCase *usecases.CheckMessages
}

func NewCheckMessages(
	logger *slog.Logger,
	useCase *usecases.CheckMessages,
) *CheckMessages {
	return &CheckMessages{
		logger:  logger,
		useCase: useCase,
	}
}

func (a *CheckMessages) Mount(srv *http.ServeMux) {
	srv.Handle("/messages/check", base.NewTypedHandler(a.logger, a.handler))
}

func (a *CheckMessages) handler(
	ctx context.Context,
	req httpmodels.CheckRequest,
) (httpmodels.CheckResponse, *base.Error) {
	result, err := a.useCase.Do(ctx, req)
	if err != nil {
		return nil, base.NewError(http.StatusInternalServerError, err)
	}

	response := make([]httpmodels.Message, 0, len(result))

	for _, msg := range result {
		history := make([]httpmodels.MessageChapter, 0, len(msg.History))

		for _, chap := range msg.History {
			history = append(history, httpmodels.MessageChapter{
				Generation:   chap.Generation,
				Queue:        chap.Queue.String(),
				RedirectedAt: chap.RedirectedAt,
				Priority:     chap.Priority,
				Retries:      chap.Retries,
			})
		}

		response = append(response, httpmodels.Message{
			ID:          msg.ID,
			Queue:       msg.Queue.String(),
			CreatedAt:   msg.CreatedAt,
			FinalizedAt: msg.FinalizedAt,
			Status:      httpmodels.MessageStatus(msg.Status),
			Priority:    msg.Priority,
			Retries:     msg.Retries,
			Generation:  msg.Generation,
			History:     history,
			Payload:     msg.Payload,
		})
	}

	return response, nil
}

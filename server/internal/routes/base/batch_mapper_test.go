package base

import (
	"testing"

	"github.com/stretchr/testify/require"

	"server/internal/usecases"
	"server/internal/utils"
	"server/pkg/httpmodels"
)

func TestCombineResultsWithMapErrors(t *testing.T) {
	const (
		resultData      = 123
		otherResultData = 231
		someError       = "some error"
		someOtherError  = "some other error"
	)

	tests := []struct {
		name      string
		results   []usecases.BatchResult[int]
		mapErrors map[int]*httpmodels.Error
		expected  []usecases.BatchResult[int]
	}{
		{
			name: "no mapping errors",
			results: []usecases.BatchResult[int]{
				{Data: utils.P(resultData)},
				{Error: httpmodels.NewError(httpmodels.ErrorCodeUnknown, someError)},
			},
			mapErrors: nil,
			expected: []usecases.BatchResult[int]{
				{Data: utils.P(resultData)},
				{Error: httpmodels.NewError(httpmodels.ErrorCodeUnknown, someError)},
			},
		},
		{
			name:    "all mapping errors",
			results: nil,
			mapErrors: map[int]*httpmodels.Error{
				0: httpmodels.NewError(httpmodels.ErrorCodeRequestInvalid, someError),
				1: httpmodels.NewError(httpmodels.ErrorCodeRequestInvalid, someOtherError),
			},
			expected: []usecases.BatchResult[int]{
				{Error: httpmodels.NewError(httpmodels.ErrorCodeRequestInvalid, someError)},
				{Error: httpmodels.NewError(httpmodels.ErrorCodeRequestInvalid, someOtherError)},
			},
		},
		{
			name: "mapping errors in the beginning",
			results: []usecases.BatchResult[int]{
				{Data: utils.P(resultData)},
			},
			mapErrors: map[int]*httpmodels.Error{
				0: httpmodels.NewError(httpmodels.ErrorCodeRequestInvalid, someError),
			},
			expected: []usecases.BatchResult[int]{
				{Error: httpmodels.NewError(httpmodels.ErrorCodeRequestInvalid, someError)},
				{Data: utils.P(resultData)},
			},
		},
		{
			name: "mapping errors in the end",
			results: []usecases.BatchResult[int]{
				{Data: utils.P(resultData)},
			},
			mapErrors: map[int]*httpmodels.Error{
				1: httpmodels.NewError(httpmodels.ErrorCodeRequestInvalid, someError),
			},
			expected: []usecases.BatchResult[int]{
				{Data: utils.P(resultData)},
				{Error: httpmodels.NewError(httpmodels.ErrorCodeRequestInvalid, someError)},
			},
		},
		{
			name: "mapping errors in the middle",
			results: []usecases.BatchResult[int]{
				{Data: utils.P(resultData)},
				{Data: utils.P(otherResultData)},
			},
			mapErrors: map[int]*httpmodels.Error{
				1: httpmodels.NewError(httpmodels.ErrorCodeRequestInvalid, someError),
			},
			expected: []usecases.BatchResult[int]{
				{Data: utils.P(resultData)},
				{Error: httpmodels.NewError(httpmodels.ErrorCodeRequestInvalid, someError)},
				{Data: utils.P(otherResultData)},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, combineResultsWithMapItemErrors(tt.results, tt.mapErrors))
		})
	}
}

package base

import (
	"errors"

	"server/internal/usecases"
	"server/pkg/httpmodels"
)

// MapBatchRequestItems maps a batch of input items to output items.
//
// If mapping of an item fails, the item is skipped and the mapping error is recorded in the
// returned error map under the original item index. The error map is intended to
// be passed to MapBatchResults so the response can preserve per-item failures.
func MapBatchRequestItems[T1, T2 any](
	items []T1,
	mapper func(T1) (T2, *httpmodels.Error),
) ([]T2, map[int]*httpmodels.Error) {
	var mapped []T2
	mapErrors := make(map[int]*httpmodels.Error)

	for i, item := range items {
		mappedItem, err := mapper(item)
		if err != nil {
			mapErrors[i] = err
			continue
		}

		mapped = append(mapped, mappedItem)
	}

	return mapped, mapErrors
}

// MapBatchResults merges batch processing results with item-mapping errors and
// converts them to API batch results.
func MapBatchResults[T1, T2 any](
	mapItemErrors map[int]*httpmodels.Error,
	results []usecases.BatchResult[T1],
	mapper func(*T1) *T2,
) []httpmodels.BatchResult[T2] {
	results = combineResultsWithMapItemErrors(results, mapItemErrors)

	mappedResults := make([]httpmodels.BatchResult[T2], len(results))

	for i, result := range results {
		if result.Error != nil {
			var mappedErr *httpmodels.Error

			// Errors from mapItemErrors are already mapped
			if !errors.As(result.Error, &mappedErr) {
				mappedErr = ExtractKnownErrors(result.Error)
			}

			mappedResults[i] = httpmodels.BatchResult[T2]{
				Error: mappedErr,
			}
		} else {
			mappedResults[i] = httpmodels.BatchResult[T2]{
				Data: mapper(result.Data),
			}
		}
	}

	return mappedResults
}

func combineResultsWithMapItemErrors[T any](
	results []usecases.BatchResult[T],
	mapItemErrors map[int]*httpmodels.Error,
) []usecases.BatchResult[T] {
	totalCount := len(results) + len(mapItemErrors)
	combined := make([]usecases.BatchResult[T], 0, totalCount)

	totalIdx := 0
	resultIdx := 0

	for totalIdx < totalCount {
		if mapError, exist := mapItemErrors[totalIdx]; exist {
			combined = append(combined, usecases.BatchResult[T]{Error: mapError})
		} else {
			combined = append(combined, results[resultIdx])
			resultIdx++
		}
		totalIdx++
	}

	return combined
}

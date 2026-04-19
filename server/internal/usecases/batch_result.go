package usecases

type BatchResult[T any] struct {
	Data  *T
	Error error
}

func mapResultToBatch[T any](result *T, err error) BatchResult[T] {
	if err != nil {
		return BatchResult[T]{
			Error: err,
		}
	}

	return BatchResult[T]{
		Data: result,
	}
}

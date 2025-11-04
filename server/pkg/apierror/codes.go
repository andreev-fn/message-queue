package apierror

type Code string

const (
	CodeUnknown         Code = "unknown"
	CodeMessageNotFound Code = "message_not_found"
	CodeQueueNotFound   Code = "queue_not_found"
)

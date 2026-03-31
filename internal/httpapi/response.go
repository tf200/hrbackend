package httpapi

type Envelope[T any] struct {
	Success bool   `json:"success"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

func OK[T any](data T, message string) Envelope[T] {
	return Envelope[T]{
		Success: true,
		Message: message,
		Data:    data,
	}
}

func Fail(message string, code string) Envelope[struct{}] {
	return Envelope[struct{}]{
		Success: false,
		Code:    code,
		Message: message,
		Data:    struct{}{},
	}
}

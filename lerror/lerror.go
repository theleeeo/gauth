package lerror

import "net/http"

type Error struct {
	message    string
	wrappedErr error
	statusCode int
}

func Status(err error) int {
	if err == nil {
		return http.StatusOK
	}

	if e, ok := err.(Error); ok {
		return e.statusCode
	}

	return http.StatusInternalServerError
}

func (e Error) Error() string {
	if e.message == "" {
		return e.wrappedErr.Error()
	}

	if e.wrappedErr == nil {
		return e.message
	}

	return e.message + ": " + e.wrappedErr.Error()
}

func (e Error) Unwrap() error {
	return e.wrappedErr
}

func (e Error) Status() int {
	return e.statusCode
}

func New(message string, status int) Error {
	return Error{
		message:    message,
		statusCode: status,
	}
}

func Wrap(err error, message string, status int) Error {
	return Error{
		message:    message,
		wrappedErr: err,
		statusCode: status,
	}
}

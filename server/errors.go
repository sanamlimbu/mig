package mig

import "net/http"

type ErrorType string

const (
	InternalServerError ErrorType = "internal-sever-error"
	InvalidInputError   ErrorType = "invalid-input"
	UnknownError        ErrorType = "unknown"
	UnauthorizedError   ErrorType = "unauthorized"
	NotFoundError       ErrorType = "not-found"
)

const (
	ErrMsgSomethingWentWrong string = "Something went wrong. Please try again later."
	ErrMsgUnableToJsonEnode  string = "Unable to json encode response."
)

type Error struct {
	msg     string
	err     error
	errType ErrorType
}

func (e Error) Error() string {
	return e.err.Error()
}

func (e Error) Friendly() string {
	return e.msg
}

func (e Error) Type() ErrorType {
	return e.errType
}

func NewError(msg string, err error, errType ErrorType) Error {
	return Error{
		msg:     msg,
		err:     err,
		errType: errType,
	}
}

func HttpErrorReply(w http.ResponseWriter, e error) {
	err, ok := e.(Error)
	if !ok {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	switch err.Type() {
	case NotFoundError:
		http.Error(w, err.msg, http.StatusNotFound)
	case InternalServerError:
		http.Error(w, err.msg, http.StatusInternalServerError)
	case InvalidInputError:
		http.Error(w, err.msg, http.StatusBadRequest)
	case UnauthorizedError:
		http.Error(w, err.msg, http.StatusUnauthorized)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

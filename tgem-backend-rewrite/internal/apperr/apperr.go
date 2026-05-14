package apperr

type Code string

const (
	CodeInvalidInput     Code = "invalid_input"
	CodeNotFound         Code = "not_found"
	CodeConflict         Code = "conflict"
	CodePermissionDenied Code = "permission_denied"
	CodeInternal         Code = "internal"
)

// No CodeUnauthenticated: authentication fails in api/middleware/authenication.go
// before any usecase runs, so usecases never need to express that state.

type Error struct {
	Code    Code
	Message string
	Cause   error
}

func (e *Error) Error() string { return e.Message }
func (e *Error) Unwrap() error { return e.Cause }

func InvalidInput(msg string, cause error) *Error {
	return &Error{Code: CodeInvalidInput, Message: msg, Cause: cause}
}

func NotFound(msg string, cause error) *Error {
	return &Error{Code: CodeNotFound, Message: msg, Cause: cause}
}

func Conflict(msg string, cause error) *Error {
	return &Error{Code: CodeConflict, Message: msg, Cause: cause}
}

func PermissionDenied(msg string) *Error {
	return &Error{Code: CodePermissionDenied, Message: msg}
}

func Internal(msg string, cause error) *Error {
	return &Error{Code: CodeInternal, Message: msg, Cause: cause}
}

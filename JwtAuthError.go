package mgboot

import "fmt"

type JwtAuthError struct {
	errno int
}

func NewJwtAuthError(errno int) JwtAuthError {
	return JwtAuthError{errno: errno}
}

func (ex JwtAuthError) Error() string {
	return fmt.Sprintf("jwt auth failed, errno: %d", ex.errno)
}

func (ex JwtAuthError) Errno() int {
	return ex.errno
}

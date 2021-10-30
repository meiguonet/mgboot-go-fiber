package mgboot

import "github.com/meiguonet/mgboot-go-fiber/enum/JwtVerifyErrno"

type jwtAuthErrorHandler struct {
}

func NewJwtAuthErrorHandler() *jwtAuthErrorHandler {
	return &jwtAuthErrorHandler{}
}

func (h *jwtAuthErrorHandler) GetErrorName() string {
	return "builtin.JwtAuthError"
}

func (h *jwtAuthErrorHandler) MatchError(err error) bool {
	if _, ok := err.(JwtAuthError); ok {
		return true
	}

	return false
}

func (h *jwtAuthErrorHandler) HandleError(err error) ResponsePayload {
	ex := err.(JwtAuthError)
	var code int
	var msg string

	switch ex.Errno() {
	case JwtVerifyErrno.NotFound:
		code = 1001
		msg = "安全令牌缺失"
	case JwtVerifyErrno.Invalid:
		code = 1002
		msg = "不是有效的安全令牌"
	case JwtVerifyErrno.Expired:
		code = 1003
		msg = "安全令牌已失效"
	}

	payload := map[string]interface{}{
		"code": code,
		"msg":  msg,
		"data": nil,
	}

	return NewJsonResponse(payload)
}

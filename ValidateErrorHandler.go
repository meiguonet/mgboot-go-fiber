package mgboot

import "github.com/meiguonet/mgboot-go-common/util/jsonx"

type validateErrorHandler struct {
}

func NewValidateErrorHandler() *validateErrorHandler {
	return &validateErrorHandler{}
}

func (h *validateErrorHandler) GetErrorName() string {
	return "builtin.ValidateError"
}

func (h *validateErrorHandler) MatchError(err error) bool {
	if _, ok := err.(ValidateError); ok {
		return true
	}

	return false
}

func (h *validateErrorHandler) HandleError(err error) ResponsePayload {
	ex := err.(ValidateError)
	code := 1006
	var msg string

	if ex.Failfast() {
		msg = ex.Error()
	} else {
		msg = jsonx.ToJson(ex.ValidateErrors())
	}

	payload := map[string]interface{}{
		"code": code,
		"msg":  msg,
		"data": nil,
	}

	return NewJsonResponse(payload)
}

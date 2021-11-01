package mgboot

import "github.com/gofiber/fiber/v2"

type rateLimitErrorHandler struct {
}

func NewRateLimitErrorHandler() *rateLimitErrorHandler {
	return &rateLimitErrorHandler{}
}

func (h *rateLimitErrorHandler) GetErrorName() string {
	return "builtin.RateLimitError"
}

func (h *rateLimitErrorHandler) MatchError(err error) bool {
	if _, ok := err.(RateLimitError); ok {
		return true
	}

	return false
}

func (h *rateLimitErrorHandler) HandleError(_ error) ResponsePayload {
	return NewHttpErrorResponse(fiber.StatusTooManyRequests)
}

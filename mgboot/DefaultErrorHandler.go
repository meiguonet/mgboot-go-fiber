package mgboot

import (
	"github.com/gofiber/fiber/v2"
	"github.com/meiguonet/mgboot-go-common/util/errorx"
)

func DefaultErrorHandler() func(ctx *fiber.Ctx, err error) error {
	return func(ctx *fiber.Ctx, err error) error {
		if ex, ok := err.(*fiber.Error); ok && ex.Code == fiber.StatusMethodNotAllowed {
			ctx.Type("html", "utf8")
			ctx.Status(fiber.StatusMethodNotAllowed ).Send([]byte{})
			return nil
		}

		handlers := ErrorHandlers()
		var handler ErrorHandler

		for _, h := range handlers {
			if h.MatchError(err) {
				handler = h
				break
			}
		}

		LogExecuteTime(ctx)
		AddPoweredBy(ctx)

		if handler == nil {
			RuntimeLogger().Error(errorx.Stacktrace(err))
			ctx.Type("html", "utf8")
			ctx.Status(fiber.StatusInternalServerError).Send([]byte{})
			return nil
		}

		if ex, ok := err.(RateLimitError); ok {
			ex.AddSpecifyHeaders(ctx)
		}

		payload := handler.HandleError(err)
		statusCode, contents := payload.GetContents()

		if statusCode >= 400 {
			ctx.Type("html", "utf8")
			ctx.Status(statusCode).Send([]byte{})
			return nil
		}

		contentType := payload.GetContentType()

		if contentType != "" {
			ctx.Set(fiber.HeaderContentType, contentType)
		}

		ctx.SendString(contents)
		return nil
	}
}


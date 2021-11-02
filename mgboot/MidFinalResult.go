package mgboot

import (
	"github.com/gofiber/fiber/v2"
	"github.com/meiguonet/mgboot-go-common/AppConf"
)

func MidFinalResult() func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		if AppConf.GetBoolean("logging.logMiddlewareRun") {
			RuntimeLogger().Info("middleware run: mgboot.MidFinalResult")
		}

		LogExecuteTime(ctx)
		AddCorsSupport(ctx)
		AddPoweredBy(ctx)
		var payload ResponsePayload

		if p, ok := ctx.Locals("ResponsePayload").(ResponsePayload); ok {
			payload = p
		}

		if payload == nil {
			ctx.Type("html", "utf8")
			ctx.SendString("unsupported response payload found")
			return nil
		}

		statusCode, contents := payload.GetContents()

		if statusCode >= 400 {
			ctx.Type("html", "utf8")
			ctx.Status(500).Send([]byte{})
			return nil
		}

		if pl, ok := payload.(AttachmentResponse); ok {
			pl.AddSpecifyHeaders(ctx)
			ctx.Send(pl.Buffer())
			return nil
		}

		if pl, ok := payload.(ImageResponse); ok {
			ctx.Set(fiber.HeaderContentType, pl.GetContentType())
			ctx.Send(pl.Buffer())
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

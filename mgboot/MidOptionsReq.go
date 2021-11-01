package mgboot

import (
	"github.com/gofiber/fiber/v2"
)

func MidOptionsReq() func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		if ctx.Method() != "OPTIONS" {
			return ctx.Next()
		}

		AddCorsSupport(ctx)
		AddPoweredBy(ctx)
		ctx.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSONCharsetUTF8)
		ctx.SendString(`{"code":200}`)
		return nil
	}
}

package mgboot

import (
	"github.com/gofiber/fiber/v2"
	"github.com/meiguonet/mgboot-go-common/util/validatex"
)

func MidValidate() func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		var req *Request

		if r, ok := ctx.Locals("request").(*Request); ok {
			req = r
		}

		if req == nil || req.ValidateSettings() == nil {
			return ctx.Next()
		}

		settings := req.ValidateSettings()
		validator := validatex.NewValidator()
		data := req.GetMap(ctx)

		if settings.Failfast() {
			errorTips := validatex.FailfastValidate(validator, data, settings.Rules())

			if errorTips != "" {
				return NewValidateError(errorTips, true)
			}

			return ctx.Next()
		}

		validateErrors := validatex.Validate(validator, data, settings.Rules())

		if len(validateErrors) > 0 {
			return NewValidateError(validateErrors)
		}

		return ctx.Next()
	}
}

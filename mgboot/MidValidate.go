package mgboot

import (
	"github.com/gofiber/fiber/v2"
	"github.com/meiguonet/mgboot-go-common/AppConf"
	"github.com/meiguonet/mgboot-go-common/util/validatex"
)

func MidValidate() func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		if AppConf.GetBoolean("logging.logMiddlewareRun") {
			RuntimeLogger().Info("middleware run: mgboot.MidValidate")
		}

		req := NewRequest(ctx)
		settings := FindMatchedValidateSettings(req)

		if settings == nil || len(settings.rules) < 1 {
			return ctx.Next()
		}

		validator := validatex.NewValidator()
		data := req.GetMap()

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

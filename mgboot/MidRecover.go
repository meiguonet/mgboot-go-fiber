package mgboot

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/meiguonet/mgboot-go-common/AppConf"
)

func MidRecover() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		if AppConf.GetBoolean("logging.logMiddlewareRun") {
			RuntimeLogger().Info("middleware run: mgboot.MidRecover")
		}

		var err error

		defer func() {
			r := recover()

			if r == nil {
				return
			}

			if ex, ok := r.(error); ok {
				err = ex
			} else {
				err = fmt.Errorf("%v", r)
			}
		}()

		err = ctx.Next()
		return err
	}
}

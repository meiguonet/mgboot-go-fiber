package mgboot

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/meiguonet/mgboot-go-common/AppConf"
	"github.com/meiguonet/mgboot-go-common/enum/RegexConst"
	"github.com/meiguonet/mgboot-go-common/util/castx"
	"github.com/meiguonet/mgboot-go-common/util/jsonx"
	"github.com/meiguonet/mgboot-go-common/util/numberx"
	"github.com/meiguonet/mgboot-go-common/util/stringx"
	"github.com/meiguonet/mgboot-go-common/util/validatex"
	"github.com/meiguonet/mgboot-go-dal/ratelimiter"
	"github.com/meiguonet/mgboot-go-fiber/enum/JwtVerifyErrno"
	"math/big"
	"mime/multipart"
	"strings"
	"time"
)

type ImageInfoGetFunc func(fh *multipart.FileHeader) map[string]interface{}

var Version = "1.2.2"
var errorHandlers = make([]ErrorHandler, 0)

func LogExecuteTime(ctx *fiber.Ctx) {
	if !ExecuteTimeLogEnabled() {
		return
	}

	elapsedTime := calcElapsedTime(ctx)

	if elapsedTime == "" {
		return
	}

	sb := strings.Builder{}
	sb.WriteString(ctx.Method())
	sb.WriteString(" ")
	sb.WriteString(GetRequestUrl(ctx, true))
	sb.WriteString(", total elapsed time: " + elapsedTime)
	ExecuteTimeLogLogger().Info(sb.String())
	ctx.Set("X-Response-Time", elapsedTime)
}

func WithBuiltinErrorHandlers() {
	errorHandlers = []ErrorHandler{
		NewRateLimitErrorHandler(),
		NewJwtAuthErrorHandler(),
		NewValidateErrorHandler(),
	}
}

func ReplaceBuiltinErrorHandler(errName string, handler ErrorHandler) {
	errName = stringx.EnsureRight(errName, "Error")
	errName = stringx.EnsureLeft(errName, "builtin.")
	handlers := make([]ErrorHandler, 0)
	var added bool

	for _, h := range errorHandlers {
		if h.GetErrorName() == errName {
			handlers = append(handlers, handler)
			added = true
			continue
		}

		handlers = append(handlers, h)
	}

	if !added {
		handlers = append(handlers, handler)
	}

	errorHandlers = handlers
}

func WithErrorHandler(handler ErrorHandler) {
	handlers := make([]ErrorHandler, 0)
	var added bool

	for _, h := range errorHandlers {
		if h.GetErrorName() == handler.GetErrorName() {
			handlers = append(handlers, handler)
			added = true
			continue
		}

		handlers = append(handlers, h)
	}

	if !added {
		handlers = append(handlers, handler)
	}

	errorHandlers = handlers
}

func WithErrorHandlers(handlers []ErrorHandler) {
	if len(handlers) < 1 {
		return
	}

	for _, handler := range handlers {
		WithErrorHandler(handler)
	}
}

func ErrorHandlers() []ErrorHandler {
	return errorHandlers
}

func AddPoweredBy(ctx *fiber.Ctx) {
	poweredBy := AppConf.GetString("app.poweredBy")

	if poweredBy == "" {
		return
	}

	ctx.Set("X-Powered-By", poweredBy)
}

func RateLimitCheck(ctx *fiber.Ctx, handlerName string, settings interface{}) error {
	var total int
	var duration time.Duration
	var limitByIp bool

	if map1, ok := settings.(map[string]interface{}); ok && len(map1) > 0 {
		total = castx.ToInt(map1["total"])

		if d1, ok := map1["duration"].(time.Duration); ok && d1 > 0 {
			duration = d1
		} else if n1, err := castx.ToInt64E(map1["duration"]); err == nil && n1 > 0 {
			duration = time.Duration(n1) * time.Millisecond
		}

		limitByIp = castx.ToBool(map1["limitByIp"])
	} else if s1, ok := settings.(string); ok && s1 != "" {
		s1 = strings.ReplaceAll(s1, "[syh]", `"`)
		map1 := jsonx.MapFrom(s1)

		if len(map1) > 0 {
			total = castx.ToInt(map1["total"])

			if d1, ok := map1["duration"].(time.Duration); ok && d1 > 0 {
				duration = d1
			} else if n1, err := castx.ToInt64E(map1["duration"]); err == nil && n1 > 0 {
				duration = time.Duration(n1) * time.Millisecond
			}

			limitByIp = castx.ToBool(map1["limitByIp"])
		}
	}

	if handlerName == "" || total < 1 || duration < 1 {
		return nil
	}

	id := handlerName

	if limitByIp {
		id += "@" + GetClientIp(ctx)
	}

	opts := ratelimiter.NewRatelimiterOptions(RatelimiterLuaFile(), RatelimiterCacheDir())
	limiter := ratelimiter.NewRatelimiter(id, total, duration, opts)
	result := limiter.GetLimit()
	remaining := castx.ToInt(result["remaining"])

	if remaining < 0 {
		return NewRateLimitError(result)
	}

	return nil
}

func JwtAuthCheck(ctx *fiber.Ctx, settingsKey string) error {
	if settingsKey == "" {
		return nil
	}

	settings := GetJwtSettings(settingsKey)

	if settings == nil {
		return nil
	}

	token := strings.TrimSpace(ctx.Get(fiber.HeaderAuthorization))
	token = stringx.RegexReplace(token, RegexConst.SpaceSep, " ")

	if strings.Contains(token, " ") {
		token = stringx.SubstringAfter(token, " ")
	}

	if token == "" {
		return NewJwtAuthError(JwtVerifyErrno.NotFound)
	}

	errno := VerifyJsonWebToken(token, settings)

	if errno < 0 {
		return NewJwtAuthError(errno)
	}

	return nil
}

func ValidateCheck(ctx *fiber.Ctx, settings interface{}) error {
	rules := make([]string, 0)
	var failfast bool

	if items, ok := settings.([]string); ok && len(items) > 0 {
		for _, s1 := range items {
			if s1 == "" || s1 == "false" {
				continue
			}

			if s1 == "true" {
				failfast = true
				continue
			}

			rules = append(rules, s1)
		}
	} else if s1, ok := settings.(string); ok && s1 != "" {
		s1 = strings.ReplaceAll(s1, "[syh]", `"`)
		entries := jsonx.ArrayFrom(s1)

		for _, entry := range entries {
			s2, ok := entry.(string)

			if !ok || s2 == "" || s2 == "false" {
				continue
			}

			if s2 == "true" {
				failfast = true
				continue
			}

			rules = append(rules, s2)
		}
	}

	if len(rules) < 1 {
		return nil
	}

	validator := validatex.NewValidator()
	data := GetMap(ctx)

	if failfast {
		errorTips := validatex.FailfastValidate(validator, data, rules)

		if errorTips != "" {
			return NewValidateError(errorTips, true)
		}

		return nil
	}

	validateErrors := validatex.Validate(validator, data, rules)

	if len(validateErrors) > 0 {
		return NewValidateError(validateErrors)
	}

	return nil
}

func CheckUploadedFile(fh *multipart.FileHeader, opts map[string]interface{}) (passed bool, errorTips string) {
	if fh == nil {
		errorTips = "没有文件被上传"
		return
	}

	var maxFileSize int64

	if s1, ok := opts["fileSizeLimit"]; ok && s1 != "" {
		maxFileSize = castx.ToDataSize(s1)
	}

	if maxFileSize > 0 && fh.Size > maxFileSize {
		errorTips = "文件大小超出限制"
		return
	}

	if !castx.ToBool(opts["checkImage"]) {
		return
	}

	var fn ImageInfoGetFunc

	if f1, ok := opts["imageInfoFunc"].(ImageInfoGetFunc); ok {
		fn = f1
	}

	if fn == nil {
		return
	}

	map1 := fn(fh)
	width := castx.ToInt(map1["width"])
	height := castx.ToInt(map1["height"])
	mimeType := castx.ToString(map1["mimeType"])

	if width < 1 || height < 1 || mimeType == "" {
		errorTips = "不是有效的图片文件"
		return
	}

	imageSizeLimit := castx.ToString(opts["imageSizeLimit"])

	if imageSizeLimit != "" {
		var n1 int
		var n2 int
		parts := stringx.SplitWithRegexp(strings.TrimSpace(imageSizeLimit), `[\x20\t]*x[\x20\t]*`)

		if len(parts) >= 2 {
			n1 = castx.ToInt(parts[0])
			n2 = castx.ToInt(parts[1])
		}

		if n1 > 0 && n2 > 0 && (width != n1 || height != n2) {
			errorTips = fmt.Sprintf("请上传%dx%d的图片", n1, n2)
			return
		}
	}

	imageRatioLimit := castx.ToString(opts["imageRatioLimit"])

	if imageRatioLimit != "" {
		var n1 int
		var n2 int
		parts := stringx.SplitWithRegexp(strings.TrimSpace(imageRatioLimit), `[\x20\t]*:[\x20\t]*`)

		if len(parts) >= 2 {
			n1 = castx.ToInt(parts[0])
			n2 = castx.ToInt(parts[1])
		}

		if n1 > 0 && n2 > 0 {
			n3 := numberx.Ojld(width, height)
			n4 := width / n3
			n5 := height / n3

			if n4 != n1 || n5 != n2 {
				errorTips = fmt.Sprintf("请上传%d:%d比例的图片", n1, n2)
				return
			}
		}
	}

	return
}

func SendOutput(ctx *fiber.Ctx, payload ResponsePayload, err error) error {
	if err != nil {
		handler := DefaultErrorHandler()
		_ = handler(ctx, err)
		return nil
	}

	LogExecuteTime(ctx)
	AddPoweredBy(ctx)

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

func calcElapsedTime(ctx *fiber.Ctx) string {
	var execStart time.Time

	if t1, ok := ctx.Locals("ExecStart").(time.Time); ok {
		ctx.Locals("ExecStart", nil)
		execStart = t1
	}

	if execStart.IsZero() {
		return ""
	}

	n1 := big.NewFloat(time.Since(execStart).Seconds())

	if n1.Cmp(big.NewFloat(1.0)) != -1 {
		secs, _ := n1.Float64()
		return numberx.ToDecimalString(secs, 3) + "s"
	}

	n1 = n1.Mul(n1, big.NewFloat(1000.0))

	if n1.Cmp(big.NewFloat(1.0)) == -1 {
		return "0ms"
	}

	msecs, _ := n1.Float64()
	return fmt.Sprintf("%dms", castx.ToInt(msecs))
}

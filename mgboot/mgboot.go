package mgboot

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/meiguonet/mgboot-go-common/AppConf"
	"github.com/meiguonet/mgboot-go-common/logx"
	"github.com/meiguonet/mgboot-go-common/util/castx"
	"github.com/meiguonet/mgboot-go-common/util/numberx"
	"github.com/meiguonet/mgboot-go-common/util/slicex"
	"github.com/meiguonet/mgboot-go-common/util/stringx"
	"strings"
	"time"
)

var runtimeLogger logx.Logger
var requestLogLogger logx.Logger
var logRequestBody bool
var executeTimeLogLogger logx.Logger
var errorHandlers = make([]ErrorHandler, 0)
var handlerNameMap = map[string]string{}
var rateLimitSettingsMap = map[string]*RateLimitSettings{}
var jwtAuthSettingsMap = map[string]*JwtAuthSettings{}
var validateSettingsMap = map[string]*ValidateSettings{}

func RuntimeLogger(logger ...logx.Logger) logx.Logger {
	if len(logger) > 0 {
		runtimeLogger = logger[0]
	}

	l := runtimeLogger

	if l == nil {
		l = NewNoopLogger()
	}

	return l
}

func RequestLogLogger(logger ...logx.Logger) logx.Logger {
	if len(logger) > 0 {
		requestLogLogger = logger[0]
	}

	l := requestLogLogger

	if l == nil {
		l = NewNoopLogger()
	}

	return l
}

func RequestLogEnabled() bool {
	return requestLogLogger != nil
}

func LogRequestBody(flag ...bool) bool {
	if len(flag) > 0 {
		logRequestBody = flag[0]
	}

	return logRequestBody
}

func ExecuteTimeLogLogger(logger ...logx.Logger) logx.Logger {
	if len(logger) > 0 {
		executeTimeLogLogger = logger[0]
	}

	l := executeTimeLogLogger

	if l == nil {
		l = NewNoopLogger()
	}

	return l
}

func ExecuteTimeLogEnabled() bool {
	return executeTimeLogLogger != nil
}

func LogExecuteTime(ctx *fiber.Ctx) {
	if !ExecuteTimeLogEnabled() {
		return
	}

	req := NewRequest(ctx)
	elapsedTime := calcElapsedTime(ctx)

	if elapsedTime == "" {
		return
	}

	handlerName := FindMatchedHandlerName(req)
	sb := strings.Builder{}
	sb.WriteString(ctx.Method())
	sb.WriteString(" ")
	sb.WriteString(req.GetRequestUrl(true))

	if handlerName != "" {
		sb.WriteString(", handler: " + handlerName)
	}

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

func NeedCorsSupport(ctx *fiber.Ctx) bool {
	req := NewRequest(ctx)
	methods := []string{"PUT", "DELETE", "CONNECT", "OPTIONS", "TRACE", "PATCH"}

	if slicex.InStringSlice(ctx.Method(), methods) {
		return true
	}

	contentType := strings.ToLower(ctx.Get(fiber.HeaderContentType))

	if strings.Contains(contentType, fiber.MIMEApplicationForm) ||
		strings.Contains(contentType, fiber.MIMEMultipartForm) ||
		strings.Contains(contentType, fiber.MIMETextPlain) {
		return true
	}

	headerNames := []string{
		"Accept",
		"Accept-Language",
		"Content-Language",
		"DPR",
		"Downlink",
		"Save-Data",
		"Viewport-Widt",
		"Width",
	}

	for headerName := range req.GetHeaders() {
		if slicex.InStringSlice(headerName, headerNames) {
			return true
		}
	}

	return false
}

func AddCorsSupport(ctx *fiber.Ctx) {
	if !NeedCorsSupport(ctx) {
		return
	}
	
	settings := GetCorsSettings()
	
	if settings == nil {
		return
	}

	allowedOrigins := settings.AllowedOrigins()

	if slicex.InStringSlice("*", allowedOrigins) {
		ctx.Set("Access-Control-Allow-Origin", "*")
	} else {
		ctx.Set("Access-Control-Allow-Origin", strings.Join(allowedOrigins, ", "))
	}

	allowedHeaders := settings.AllowedHeaders()

	if len(allowedHeaders) > 0 {
		ctx.Set("Access-Control-Allow-Headers", strings.Join(allowedHeaders, ", "))
	}

	exposedHeaders := settings.ExposedHeaders()

	if len(exposedHeaders) > 0 {
		ctx.Set("Access-Control-Expose-Headers", strings.Join(exposedHeaders, ", "))
	}

	maxAge := settings.MaxAge()

	if maxAge > 0 {
		n1 := castx.ToInt64(maxAge.Seconds())
		ctx.Set("Access-Control-Max-Age", fmt.Sprintf("%d", n1))
	}

	if settings.AllowCredentials() {
		ctx.Set("Access-Control-Allow-Credentials", "true")
	}
}

func AddPoweredBy(ctx *fiber.Ctx) {
	poweredBy := AppConf.GetString("app.poweredBy")

	if poweredBy == "" {
		return
	}

	ctx.Set("X-Powered-By", poweredBy)
}

func WithHandlerName(method, requestMapping, handlerName string) {
	requestMapping = getRealRequestMapping(requestMapping)

	if method == "ALL" {
		handlerNameMap["GET@" + requestMapping] = handlerName
		handlerNameMap["POST@" + requestMapping] = handlerName
	} else {
		handlerNameMap[method + "@" + requestMapping] = handlerName
	}
}

func FindMatchedHandlerName(req *Request) string {
	for key, value := range handlerNameMap {
		method := stringx.SubstringBefore(key, "@")

		if method != req.GetMethod() {
			continue
		}

		requestMapping := stringx.SubstringAfter(key, "@")

		if strings.HasPrefix(requestMapping, "^") {
			continue
		}

		if requestMapping == req.GetRequestUrl() {
			return value
		}
	}

	for key, value := range handlerNameMap {
		method := stringx.SubstringBefore(key, "@")

		if method != req.GetMethod() {
			continue
		}

		requestMapping := stringx.SubstringAfter(key, "@")

		if !strings.HasPrefix(requestMapping, "^") {
			continue
		}

		if stringx.RegexMatch(req.GetRequestUrl(), requestMapping) {
			return value
		}
	}

	return ""
}

func WithRateLimitSettings(method, requestMapping string, settings interface{}) {
	requestMapping = getRealRequestMapping(requestMapping)
	st := NewRateLimitSettings(settings)

	if method == "ALL" {
		rateLimitSettingsMap["GET@" + requestMapping] = st
		rateLimitSettingsMap["POST@" + requestMapping] = st
	} else {
		rateLimitSettingsMap[method + "@" + requestMapping] = st
	}
}

func FindMatchedRateLimitSettings(req *Request) *RateLimitSettings {
	for key, value := range rateLimitSettingsMap {
		method := stringx.SubstringBefore(key, "@")

		if method != req.GetMethod() {
			continue
		}

		requestMapping := stringx.SubstringAfter(key, "@")

		if strings.HasPrefix(requestMapping, "^") {
			continue
		}

		if requestMapping == req.GetRequestUrl() {
			return value
		}
	}

	for key, value := range rateLimitSettingsMap {
		method := stringx.SubstringBefore(key, "@")

		if method != req.GetMethod() {
			continue
		}

		requestMapping := stringx.SubstringAfter(key, "@")

		if !strings.HasPrefix(requestMapping, "^") {
			continue
		}

		if stringx.RegexMatch(req.GetRequestUrl(), requestMapping) {
			return value
		}
	}

	return nil
}

func WithJwtAuthSettings(method, requestMapping, settingsKey string) {
	requestMapping = getRealRequestMapping(requestMapping)
	st := NewJwtAuthSettings(settingsKey)

	if method == "ALL" {
		jwtAuthSettingsMap["GET@" + requestMapping] = st
		jwtAuthSettingsMap["POST@" + requestMapping] = st
	} else {
		jwtAuthSettingsMap[method + "@" + requestMapping] = st
	}
}

func FindMatchedJwtAuthSettings(req *Request) *JwtAuthSettings {
	for key, value := range jwtAuthSettingsMap {
		method := stringx.SubstringBefore(key, "@")

		if method != req.GetMethod() {
			continue
		}

		requestMapping := stringx.SubstringAfter(key, "@")

		if strings.HasPrefix(requestMapping, "^") {
			continue
		}

		if requestMapping == req.GetRequestUrl() {
			return value
		}
	}

	for key, value := range jwtAuthSettingsMap {
		method := stringx.SubstringBefore(key, "@")

		if method != req.GetMethod() {
			continue
		}

		requestMapping := stringx.SubstringAfter(key, "@")

		if !strings.HasPrefix(requestMapping, "^") {
			continue
		}

		if stringx.RegexMatch(req.GetRequestUrl(), requestMapping) {
			return value
		}
	}

	return nil
}

func WithValidateSettings(method, requestMapping string, settings interface{}) {
	requestMapping = getRealRequestMapping(requestMapping)
	st := NewValidateSettings(settings)

	if method == "ALL" {
		validateSettingsMap["GET@" + requestMapping] = st
		validateSettingsMap["POST@" + requestMapping] = st
	} else {
		validateSettingsMap[method + "@" + requestMapping] = st
	}
}

func FindMatchedValidateSettings(req *Request) *ValidateSettings {
	for key, value := range validateSettingsMap {
		method := stringx.SubstringBefore(key, "@")

		if method != req.GetMethod() {
			continue
		}

		requestMapping := stringx.SubstringAfter(key, "@")

		if strings.HasPrefix(requestMapping, "^") {
			continue
		}

		if requestMapping == req.GetRequestUrl() {
			return value
		}
	}

	for key, value := range validateSettingsMap {
		method := stringx.SubstringBefore(key, "@")

		if method != req.GetMethod() {
			continue
		}

		requestMapping := stringx.SubstringAfter(key, "@")

		if !strings.HasPrefix(requestMapping, "^") {
			continue
		}

		if stringx.RegexMatch(req.GetRequestUrl(), requestMapping) {
			return value
		}
	}

	return nil
}

func calcElapsedTime(ctx *fiber.Ctx) string {
	var execStart time.Time

	if d1, ok := ctx.Locals("ExecStart").(time.Time); ok {
		execStart = d1
	}

	if execStart.IsZero() {
		return ""
	}

	d2 := time.Now().Sub(execStart)

	if d2 < time.Second {
		return fmt.Sprintf("%dms", d2)
	}

	n1 := d2.Seconds()
	return numberx.ToDecimalString(n1, 3) + "ms"
}

func getRealRequestMapping(requestMapping string) string {
	if !strings.Contains(requestMapping, ":") {
		return requestMapping
	}

	parts := strings.Split(strings.Trim(requestMapping, "/"), "/")
	sb := strings.Builder{}
	n1 := 0

	for _, p := range parts {
		if n1 > 0 {
			sb.WriteString("/")
		}

		if !strings.Contains(p, ":") {
			sb.WriteString(p)
			n1++
			continue
		}

		sb.WriteString(`[^/]+`)
		n1++
	}

	return "^/" + sb.String() + "$"
}

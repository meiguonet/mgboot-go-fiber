package mgboot

import (
	"bufio"
	"bytes"
	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/meiguonet/mgboot-go-common/AppConf"
	"github.com/meiguonet/mgboot-go-common/enum/RegexConst"
	"github.com/meiguonet/mgboot-go-common/util/castx"
	"github.com/meiguonet/mgboot-go-common/util/jsonx"
	"github.com/meiguonet/mgboot-go-common/util/mapx"
	"github.com/meiguonet/mgboot-go-common/util/slicex"
	"github.com/meiguonet/mgboot-go-common/util/stringx"
	"math"
	"net/url"
	"regexp"
	"strings"
)

type Request struct {
	handlerFuncName   string
	rateLimitSettings *RateLimitSettings
	jwtAuthSettings   *JwtAuthSettings
	validateSettings  *ValidateSettings
}

func NewRequest(settings map[string]interface{}) *Request {
	var handlerFuncName string

	if s1, ok := settings["handlerFuncName"].(string); ok {
		handlerFuncName = s1
	}

	var rateLimitSettings *RateLimitSettings

	if st, ok := settings["rateLimitSettings"].(*RateLimitSettings); ok {
		rateLimitSettings = st
	}

	var jwtAuthSettings *JwtAuthSettings

	if st, ok := settings["jwtAuthSettings"].(*JwtAuthSettings); ok {
		jwtAuthSettings = st
	}

	var validateSettings *ValidateSettings

	if st, ok := settings["validateSettings"].(*ValidateSettings); ok {
		validateSettings = st
	}

	return &Request{
		handlerFuncName:   handlerFuncName,
		rateLimitSettings: rateLimitSettings,
		jwtAuthSettings:   jwtAuthSettings,
		validateSettings:  validateSettings,
	}
}

func (r *Request) HandlerFuncName() string {
	return r.handlerFuncName
}

func (r *Request) RateLimitSettings() *RateLimitSettings {
	return r.rateLimitSettings
}

func (r *Request) JwtAuthSettings() *JwtAuthSettings {
	return r.jwtAuthSettings
}

func (r *Request) ValidateSettings() *ValidateSettings {
	return r.validateSettings
}

func (r *Request) GetMethod(ctx *fiber.Ctx) string {
	return ctx.Method()
}

func (r *Request) GetHeaders(ctx *fiber.Ctx) map[string]string {
	if AppConf.GetBoolean("logging.logGetHeaders") {
		RuntimeLogger().Debug("raw headers: " + string(ctx.Request().Header.RawHeaders()))
	}

	buf := make([]byte, 0, len(ctx.Request().Header.RawHeaders()))
	copy(ctx.Request().Header.RawHeaders(), buf)
	reader := bufio.NewReader(bytes.NewReader(buf))
	headers := map[string]string{}

	for {
		line, _, err := reader.ReadLine()

		if err != nil {
			break
		}

		s1 := string(line)

		if s1 == "" || !strings.Contains(s1, ":") {
			continue
		}

		headerName := strings.TrimSpace(stringx.SubstringBefore(s1, ":"))
		headerValue := strings.TrimSpace(stringx.SubstringAfter(s1, ":"))

		if headerName == "" || headerValue == "" {
			continue
		}

		headerName = stringx.Ucwords(headerName, "-", "-")
		headers[headerName] = headerValue
	}

	return headers
}

func (r *Request) GetHeader(ctx *fiber.Ctx, name string) string {
	return ctx.Get(name)
}

func (r *Request) GetQueryParams(ctx *fiber.Ctx) map[string]string {
	map1 := map[string]string{}
	buf := ctx.Request().URI().QueryString()

	if len(buf) < 1 {
		return map1
	}

	if AppConf.GetBoolean("logging.logGetQueryParams") {
		RuntimeLogger().Debug("query params: " + string(buf))
	}

	parts := strings.Split(string(buf), "&")

	for _, p := range parts {
		if !strings.Contains(p, "=") {
			continue
		}

		name := stringx.SubstringBefore(p, "=")
		value, err := url.QueryUnescape(stringx.SubstringAfter(p, "+"))

		if err != nil {
			continue
		}

		map1[name] = value
	}

	return map1
}

func (r *Request) GetQueryString(ctx *fiber.Ctx, urlencode ...bool) string {
	params := r.GetQueryParams(ctx)

	if len(params) < 1 {
		return ""
	}

	if len(urlencode) > 0 && urlencode[0] {
		values := url.Values{}

		for name, value := range params {
			values[name] = []string{value}
		}

		return values.Encode()
	}

	sb := strings.Builder{}
	n1 := 0

	for name, value := range params {
		sb.WriteString(name + "=" + value)

		if n1 > 0 {
			sb.WriteString("&")
		}

		n1++
	}

	return sb.String()
}

func (r *Request) GetRequestUrl(ctx *fiber.Ctx, withQueryString ...bool) string {
	buf := ctx.Request().URI().Path()
	var s1 string

	if len(buf) < 1 {
		s1 = "/"
	} else {
		s1 = stringx.EnsureLeft(string(buf), "/")
	}

	if len(withQueryString) > 0 && withQueryString[0] {
		qs := r.GetQueryString(ctx)

		if qs != "" {
			s1 += "?" + qs
		}
	}

	return s1
}

func (r *Request) GetFormData(ctx *fiber.Ctx) map[string]string {
	map1 := map[string]string{}
	contentType := ctx.Get(fiber.HeaderContentType)
	contentTypes := []string{fiber.MIMEApplicationForm, fiber.MIMEMultipartForm}

	if ctx.Method() != "POST" || !slicex.InStringSlice(contentType, contentTypes) {
		return map1
	}

	form, err := ctx.MultipartForm()

	if err != nil {
		return map1
	}

	if AppConf.GetBoolean("logging.logGetFormData") {
		RuntimeLogger().Debug("form data: " + jsonx.ToJson(form.Value))
	}

	for name, values := range form.Value {
		if len(values) < 1 {
			continue
		}

		map1[name] = values[0]
	}

	return map1
}

func (r *Request) GetClientIp(ctx *fiber.Ctx) string {
	ip := ctx.Get(fiber.HeaderXForwardedFor)

	if ip == "" {
		ip = ctx.Get("X-Real-IP")
	}

	if ip == "" {
		ip = ctx.IP()
	}

	regex1 := regexp.MustCompile(RegexConst.CommaSep)
	parts := regex1.Split(strings.TrimSpace(ip), -1)

	if len(parts) < 1 {
		return ""
	}

	return strings.TrimSpace(parts[0])
}

func (r *Request) PathvariableString(ctx *fiber.Ctx, name string, defaultValue ...string) string {
	var dv string

	if len(defaultValue) > 0 {
		dv = defaultValue[0]
	}

	return ctx.Params(name, dv)
}

func (r *Request) PathvariableBool(ctx *fiber.Ctx, name string, defaultValue ...string) bool {
	var dv bool

	if len(defaultValue) > 0 {
		if b1, err := castx.ToBoolE(defaultValue[0]); err == nil {
			dv = b1
		}
	}

	if b1, err := castx.ToBoolE(ctx.Params(name)); err == nil {
		return b1
	}

	return dv
}

func (r *Request) PathvariableInt(ctx *fiber.Ctx, name string, defaultValue ...interface{}) int {
	dv := math.MinInt32

	if len(defaultValue) > 0 {
		if n1, err := castx.ToIntE(defaultValue[0]); err == nil {
			dv = n1
		}
	}

	return castx.ToInt(ctx.Params(name), dv)
}

func (r *Request) PathvariableInt64(ctx *fiber.Ctx, name string, defaultValue ...interface{}) int64 {
	dv := int64(math.MinInt64)

	if len(defaultValue) > 0 {
		if n1, err := castx.ToInt64E(defaultValue[0]); err == nil {
			dv = n1
		}
	}

	return castx.ToInt64(ctx.Params(name), dv)
}

func (r *Request) PathvariableFloat32(ctx *fiber.Ctx, name string, defaultValue ...interface{}) float32 {
	dv := float32(math.SmallestNonzeroFloat32)

	if len(defaultValue) > 0 {
		if n1, err := castx.ToFloat32E(defaultValue[0]); err == nil {
			dv = n1
		}
	}

	return castx.ToFloat32(ctx.Params(name), dv)
}

func (r *Request) PathvariableFloat64(ctx *fiber.Ctx, name string, defaultValue ...interface{}) float64 {
	dv := math.SmallestNonzeroFloat64

	if len(defaultValue) > 0 {
		if n1, err := castx.ToFloat64E(defaultValue[0]); err == nil {
			dv = n1
		}
	}

	return castx.ToFloat64(ctx.Params(name), dv)
}

func (r *Request) ParamString(ctx *fiber.Ctx, name string, defaultValue ...string) string {
	var dv string

	if len(defaultValue) > 0 {
		dv = defaultValue[0]
	}

	map1 := map[string]string{}

	for name, value := range r.GetQueryParams(ctx) {
		map1[name] = value
	}

	for name, value := range r.GetFormData(ctx) {
		map1[name] = value
	}

	if s1, err := castx.ToStringE(map1[name]); err == nil {
		return s1
	}

	return dv
}

func (r *Request) ParamBool(ctx *fiber.Ctx, name string, defaultValue ...string) bool {
	var dv bool

	if len(defaultValue) > 0 {
		if b1, err := castx.ToBoolE(defaultValue[0]); err == nil {
			dv = b1
		}
	}

	map1 := map[string]string{}

	for name, value := range r.GetQueryParams(ctx) {
		map1[name] = value
	}

	for name, value := range r.GetFormData(ctx) {
		map1[name] = value
	}

	if b1, err := castx.ToBoolE(map1[name]); err == nil {
		return b1
	}

	return dv
}

func (r *Request) ParamInt(ctx *fiber.Ctx, name string, defaultValue ...interface{}) int {
	dv := math.MinInt32

	if len(defaultValue) > 0 {
		if n1, err := castx.ToIntE(defaultValue[0]); err == nil {
			dv = n1
		}
	}

	map1 := map[string]string{}

	for name, value := range r.GetQueryParams(ctx) {
		map1[name] = value
	}

	for name, value := range r.GetFormData(ctx) {
		map1[name] = value
	}

	return castx.ToInt(map1[name], dv)
}

func (r *Request) ParamInt64(ctx *fiber.Ctx, name string, defaultValue ...interface{}) int64 {
	dv := int64(math.MinInt64)

	if len(defaultValue) > 0 {
		if n1, err := castx.ToInt64E(defaultValue[0]); err == nil {
			dv = n1
		}
	}

	map1 := map[string]string{}

	for name, value := range r.GetQueryParams(ctx) {
		map1[name] = value
	}

	for name, value := range r.GetFormData(ctx) {
		map1[name] = value
	}

	return castx.ToInt64(map1[name], dv)
}

func (r *Request) ParamFloat32(ctx *fiber.Ctx, name string, defaultValue ...interface{}) float32 {
	dv := float32(math.SmallestNonzeroFloat32)

	if len(defaultValue) > 0 {
		if n1, err := castx.ToFloat32E(defaultValue[0]); err == nil {
			dv = n1
		}
	}

	map1 := map[string]string{}

	for name, value := range r.GetQueryParams(ctx) {
		map1[name] = value
	}

	for name, value := range r.GetFormData(ctx) {
		map1[name] = value
	}

	return castx.ToFloat32(map1[name], dv)
}

func (r *Request) ParamFloat64(ctx *fiber.Ctx, name string, defaultValue ...interface{}) float64 {
	dv := math.SmallestNonzeroFloat64

	if len(defaultValue) > 0 {
		if n1, err := castx.ToFloat64E(defaultValue[0]); err == nil {
			dv = n1
		}
	}

	map1 := map[string]string{}

	for name, value := range r.GetQueryParams(ctx) {
		map1[name] = value
	}

	for name, value := range r.GetFormData(ctx) {
		map1[name] = value
	}

	return castx.ToFloat64(map1[name], dv)
}

func (r *Request) GetJwt(ctx *fiber.Ctx) *jwt.Token {
	token := strings.TrimSpace(ctx.Get(fiber.HeaderAuthorization))
	re := regexp.MustCompile(`[\x20\t]+`)
	token = re.ReplaceAllString(token, " ")

	if strings.Contains(token, " ") {
		token = stringx.SubstringAfter(token, " ")
	}

	if token == "" {
		return nil
	}

	tk, _ := ParseJsonWebToken(token)
	return tk
}

func (r *Request) JwtClaimString(ctx *fiber.Ctx, name string, defaultValue ...string) string {
	var dv string

	if len(defaultValue) > 0 {
		dv = defaultValue[0]
	}

	token := strings.TrimSpace(ctx.Get(fiber.HeaderAuthorization))
	re := regexp.MustCompile(`[\x20\t]+`)
	token = re.ReplaceAllString(token, " ")

	if strings.Contains(token, " ") {
		token = stringx.SubstringAfter(token, " ")
	}

	if token == "" {
		return dv
	}

	return JwtClaimString(token, name, dv)
}

func (r *Request) JwtClaimBool(ctx *fiber.Ctx, name string, defaultValue ...string) bool {
	var dv bool

	if len(defaultValue) > 0 {
		if b1, err := castx.ToBoolE(defaultValue[0]); err == nil {
			dv = b1
		}
	}

	token := strings.TrimSpace(ctx.Get(fiber.HeaderAuthorization))
	re := regexp.MustCompile(`[\x20\t]+`)
	token = re.ReplaceAllString(token, " ")

	if strings.Contains(token, " ") {
		token = stringx.SubstringAfter(token, " ")
	}

	if token == "" {
		return dv
	}

	return JwtClaimBool(token, name, dv)
}

func (r *Request) JwtClaimInt(ctx *fiber.Ctx, name string, defaultValue ...interface{}) int {
	dv := math.MinInt32

	if len(defaultValue) > 0 {
		if n1, err := castx.ToIntE(defaultValue[0]); err == nil {
			dv = n1
		}
	}

	token := strings.TrimSpace(ctx.Get(fiber.HeaderAuthorization))
	re := regexp.MustCompile(`[\x20\t]+`)
	token = re.ReplaceAllString(token, " ")

	if strings.Contains(token, " ") {
		token = stringx.SubstringAfter(token, " ")
	}

	if token == "" {
		return dv
	}

	return JwtClaimInt(token, name, dv)
}

func (r *Request) JwtClaimInt64(ctx *fiber.Ctx, name string, defaultValue ...interface{}) int64 {
	dv := int64(math.MinInt64)

	if len(defaultValue) > 0 {
		if n1, err := castx.ToInt64E(defaultValue[0]); err == nil {
			dv = n1
		}
	}

	token := strings.TrimSpace(ctx.Get(fiber.HeaderAuthorization))
	re := regexp.MustCompile(`[\x20\t]+`)
	token = re.ReplaceAllString(token, " ")

	if strings.Contains(token, " ") {
		token = stringx.SubstringAfter(token, " ")
	}

	if token == "" {
		return dv
	}

	return JwtClaimInt64(token, name, dv)
}

func (r *Request) JwtClaimFloat32(ctx *fiber.Ctx, name string, defaultValue ...interface{}) float32 {
	dv := float32(math.SmallestNonzeroFloat32)

	if len(defaultValue) > 0 {
		if n1, err := castx.ToFloat32E(defaultValue[0]); err == nil {
			dv = n1
		}
	}

	token := strings.TrimSpace(ctx.Get(fiber.HeaderAuthorization))
	re := regexp.MustCompile(`[\x20\t]+`)
	token = re.ReplaceAllString(token, " ")

	if strings.Contains(token, " ") {
		token = stringx.SubstringAfter(token, " ")
	}

	if token == "" {
		return dv
	}

	return JwtClaimFloat32(token, name, dv)
}

func (r *Request) JwtClaimFloat64(ctx *fiber.Ctx, name string, defaultValue ...interface{}) float64 {
	dv := math.SmallestNonzeroFloat64

	if len(defaultValue) > 0 {
		if n1, err := castx.ToFloat64E(defaultValue[0]); err == nil {
			dv = n1
		}
	}

	token := strings.TrimSpace(ctx.Get(fiber.HeaderAuthorization))
	re := regexp.MustCompile(`[\x20\t]+`)
	token = re.ReplaceAllString(token, " ")

	if strings.Contains(token, " ") {
		token = stringx.SubstringAfter(token, " ")
	}

	if token == "" {
		return dv
	}

	return JwtClaimFloat64(token, name, dv)
}

func (r *Request) JwtClaimStringSlice(ctx *fiber.Ctx, name string) []string {
	token := strings.TrimSpace(ctx.Get(fiber.HeaderAuthorization))
	re := regexp.MustCompile(`[\x20\t]+`)
	token = re.ReplaceAllString(token, " ")

	if strings.Contains(token, " ") {
		token = stringx.SubstringAfter(token, " ")
	}

	if token == "" {
		return make([]string, 0)
	}

	return JwtClaimStringSlice(token, name)
}

func (r *Request) JwtClaimIntSlice(ctx *fiber.Ctx, name string) []int {
	token := strings.TrimSpace(ctx.Get(fiber.HeaderAuthorization))
	re := regexp.MustCompile(`[\x20\t]+`)
	token = re.ReplaceAllString(token, " ")

	if strings.Contains(token, " ") {
		token = stringx.SubstringAfter(token, " ")
	}

	if token == "" {
		return make([]int, 0)
	}

	return JwtClaimIntSlice(token, name)
}

func (r *Request) GetRawBody(ctx *fiber.Ctx) []byte {
	method := ctx.Method()
	contentType := strings.ToLower(r.GetHeader(ctx, fiber.HeaderContentType))
	isPostForm := strings.Contains(contentType, fiber.MIMEApplicationForm)
	isMultipartForm := strings.Contains(contentType, fiber.MIMEMultipartForm)

	if method == "POST" && (isPostForm || isMultipartForm) {
		formData := r.GetFormData(ctx)

		if len(formData) < 1 {
			return make([]byte, 0)
		}

		values := url.Values{}

		for name, value := range formData {
			values[name] = []string{value}
		}

		contents := values.Encode()

		if AppConf.GetBoolean("logging.logGetRawBody") {
			RuntimeLogger().Debug("raw body: " + contents)
		}

		return []byte(contents)
	}

	methods := []string{"POST", "PUT", "PATCH", "DELETE"}

	if !slicex.InStringSlice(method, methods) {
		return make([]byte, 0)
	}

	isJson := strings.Contains(contentType, fiber.MIMEApplicationJSON)
	isXml1 := strings.Contains(contentType, fiber.MIMEApplicationXML)
	isXml2 := strings.Contains(contentType, fiber.MIMETextXML)

	if !isJson && !isXml1 && !isXml2 {
		return make([]byte, 0)
	}

	var err error
	var encoding string
	var body []byte

	ctx.Request().Header.VisitAll(func(key, value []byte) {
		if utils.UnsafeString(key) == fiber.HeaderContentEncoding {
			encoding = utils.UnsafeString(value)
		}
	})

	switch encoding {
	case fiber.StrGzip:
		body, err = ctx.Request().BodyGunzip()
	case fiber.StrBr, fiber.StrBrotli:
		body, err = ctx.Request().BodyUnbrotli()
	case fiber.StrDeflate:
		body, err = ctx.Request().BodyInflate()
	default:
		body = ctx.Request().Body()
	}

	if err != nil || len(body) < 1 {
		return make([]byte, 0)
	}

	buf := make([]byte, 0, len(body))
	copy(body, buf)

	if AppConf.GetBoolean("logging.logGetRawBody") {
		RuntimeLogger().Debug("raw body: " + string(buf))
	}

	return buf
}

// @param string[]|string rules
func (r *Request) GetMap(ctx *fiber.Ctx, rules ...interface{}) map[string]interface{} {
	method := ctx.Method()
	methods := []string{"POST", "PUT", "PATCH", "DELETE"}
	contentType := strings.ToLower(ctx.Get(fiber.HeaderContentType))
	isPostForm := strings.Contains(contentType, fiber.MIMEApplicationForm)
	isMultipartForm := strings.Contains(contentType, fiber.MIMEMultipartForm)
	isJson := strings.Contains(contentType, fiber.MIMEApplicationJSON)
	isXml1 := strings.Contains(contentType, fiber.MIMEApplicationXML)
	isXml2 := strings.Contains(contentType, fiber.MIMETextXML)
	map1 := map[string]interface{}{}

	if method == "GET" {
		for key, value := range r.GetQueryParams(ctx) {
			map1[key] = value
		}
	} else if method == "POST" && (isPostForm || isMultipartForm) {
		for key, value := range r.GetQueryParams(ctx) {
			map1[key] = value
		}

		for key, value := range r.GetFormData(ctx) {
			map1[key] = value
		}
	} else if slicex.InStringSlice(method, methods) {
		return map1
	} else if isJson {
		map1 = jsonx.MapFrom(r.GetRawBody(ctx))
	} else if isXml1 || isXml2 {
		map2 := mapx.FromXml(r.GetRawBody(ctx))

		for key, value := range map2 {
			map1[key] = value
		}
	}

	if len(map1) < 1 {
		return map[string]interface{}{}
	}

	if len(rules) < 1 {
		return map1
	}

	return mapx.FromRequestParam(map1, rules...)
}

func (r *Request) DtoBind(ctx *fiber.Ctx, dto interface{}) error {
	return mapx.BindToDto(r.GetMap(ctx), dto)
}

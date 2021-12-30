package mgboot

import (
	"bufio"
	"bytes"
	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"github.com/meiguonet/mgboot-go-common/AppConf"
	"github.com/meiguonet/mgboot-go-common/enum/RegexConst"
	"github.com/meiguonet/mgboot-go-common/util/castx"
	"github.com/meiguonet/mgboot-go-common/util/jsonx"
	"github.com/meiguonet/mgboot-go-common/util/mapx"
	"github.com/meiguonet/mgboot-go-common/util/slicex"
	"github.com/meiguonet/mgboot-go-common/util/stringx"
	"math"
	"mime/multipart"
	"net/url"
	"regexp"
	"strings"
)

type Request struct {
	ctx *fiber.Ctx
}

func NewRequest(ctx *fiber.Ctx) *Request {
	return &Request{ctx: ctx}
}

func (r *Request) GetMethod() string {
	return r.ctx.Method()
}

func (r *Request) GetHeaders() map[string]string {
	if AppConf.GetBoolean("logging.logGetHeaders") {
		RuntimeLogger().Debug("raw headers: " + string(r.ctx.Request().Header.RawHeaders()))
	}

	buf := make([]byte, 0, len(r.ctx.Request().Header.RawHeaders()))
	copy(r.ctx.Request().Header.RawHeaders(), buf)
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

func (r *Request) GetHeader(name string) string {
	return r.ctx.Get(name)
}

func (r *Request) GetQueryParams() map[string]string {
	map1 := map[string]string{}
	buf := r.ctx.Request().URI().QueryString()

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

func (r *Request) GetQueryString(urlencode ...bool) string {
	params := r.GetQueryParams()

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

func (r *Request) GetRequestUrl(withQueryString ...bool) string {
	s1 := r.ctx.Request().URI().String()

	if s1 == "" {
		s1 = "/"
	} else {
		s1 = stringx.EnsureLeft(s1, "/")
	}

	if len(withQueryString) > 0 && withQueryString[0] {
		qs := r.GetQueryString()

		if qs != "" {
			s1 += "?" + qs
		}
	}

	return s1
}

func (r *Request) GetFormData() map[string]string {
	map1 := map[string]string{}
	contentType := r.ctx.Get(fiber.HeaderContentType)
	contentTypes := []string{fiber.MIMEApplicationForm, fiber.MIMEMultipartForm}

	if r.ctx.Method() != "POST" || !slicex.InStringSlice(contentType, contentTypes) {
		return map1
	}

	form, err := r.ctx.MultipartForm()

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

func (r *Request) GetClientIp() string {
	ip := r.ctx.Get(fiber.HeaderXForwardedFor)

	if ip == "" {
		ip = r.ctx.Get("X-Real-IP")
	}

	if ip == "" {
		ip = r.ctx.IP()
	}

	regex1 := regexp.MustCompile(RegexConst.CommaSep)
	parts := regex1.Split(strings.TrimSpace(ip), -1)

	if len(parts) < 1 {
		return ""
	}

	return strings.TrimSpace(parts[0])
}

func (r *Request) PathvariableString(name string, defaultValue ...interface{}) string {
	var dv string

	if len(defaultValue) > 0 {
		if s1, err := castx.ToStringE(defaultValue[0]); err == nil {
			dv = s1
		}
	}

	return r.ctx.Params(name, dv)
}

func (r *Request) PathvariableBool(name string, defaultValue ...interface{}) bool {
	var dv bool

	if len(defaultValue) > 0 {
		if b1, err := castx.ToBoolE(defaultValue[0]); err == nil {
			dv = b1
		}
	}

	if b1, err := castx.ToBoolE(r.ctx.Params(name)); err == nil {
		return b1
	}

	return dv
}

func (r *Request) PathvariableInt(name string, defaultValue ...interface{}) int {
	dv := math.MinInt32

	if len(defaultValue) > 0 {
		if n1, err := castx.ToIntE(defaultValue[0]); err == nil {
			dv = n1
		}
	}

	return castx.ToInt(r.ctx.Params(name), dv)
}

func (r *Request) PathvariableInt64(name string, defaultValue ...interface{}) int64 {
	dv := int64(math.MinInt64)

	if len(defaultValue) > 0 {
		if n1, err := castx.ToInt64E(defaultValue[0]); err == nil {
			dv = n1
		}
	}

	return castx.ToInt64(r.ctx.Params(name), dv)
}

func (r *Request) PathvariableFloat32(name string, defaultValue ...interface{}) float32 {
	dv := float32(math.SmallestNonzeroFloat32)

	if len(defaultValue) > 0 {
		if n1, err := castx.ToFloat32E(defaultValue[0]); err == nil {
			dv = n1
		}
	}

	return castx.ToFloat32(r.ctx.Params(name), dv)
}

func (r *Request) PathvariableFloat64(name string, defaultValue ...interface{}) float64 {
	dv := math.SmallestNonzeroFloat64

	if len(defaultValue) > 0 {
		if n1, err := castx.ToFloat64E(defaultValue[0]); err == nil {
			dv = n1
		}
	}

	return castx.ToFloat64(r.ctx.Params(name), dv)
}

func (r *Request) ParamString(name string, defaultValue ...interface{}) string {
	var dv string

	if len(defaultValue) > 0 {
		if s1, err := castx.ToStringE(defaultValue[0]); err == nil {
			dv = s1
		}
	}

	map1 := map[string]string{}

	for name, value := range r.GetQueryParams() {
		map1[name] = value
	}

	for name, value := range r.GetFormData() {
		map1[name] = value
	}

	if s1, err := castx.ToStringE(map1[name]); err == nil {
		return s1
	}

	return dv
}

func (r *Request) ParamStringWithSecurityMode(name string, mode int, defaultValue ...interface{}) string {
	var dv string

	if len(defaultValue) > 0 {
		if s1, err := castx.ToStringE(defaultValue[0]); err == nil {
			dv = s1
		}
	}

	modes := []int{0, 1, 2}

	if !slicex.InIntSlice(mode, modes) {
		mode = 2
	}

	map1 := map[string]string{}

	for name, value := range r.GetQueryParams() {
		map1[name] = value
	}

	for name, value := range r.GetFormData() {
		map1[name] = value
	}

	if s1, err := castx.ToStringE(map1[name]); err == nil {
		switch mode {
		case 1, 2:
			s1 = stringx.StripTags(s1)
		}

		return s1
	}

	return dv
}

func (r *Request) ParamBool(name string, defaultValue ...interface{}) bool {
	var dv bool

	if len(defaultValue) > 0 {
		if b1, err := castx.ToBoolE(defaultValue[0]); err == nil {
			dv = b1
		}
	}

	map1 := map[string]string{}

	for name, value := range r.GetQueryParams() {
		map1[name] = value
	}

	for name, value := range r.GetFormData() {
		map1[name] = value
	}

	if b1, err := castx.ToBoolE(map1[name]); err == nil {
		return b1
	}

	return dv
}

func (r *Request) ParamInt(name string, defaultValue ...interface{}) int {
	dv := math.MinInt32

	if len(defaultValue) > 0 {
		if n1, err := castx.ToIntE(defaultValue[0]); err == nil {
			dv = n1
		}
	}

	map1 := map[string]string{}

	for name, value := range r.GetQueryParams() {
		map1[name] = value
	}

	for name, value := range r.GetFormData() {
		map1[name] = value
	}

	return castx.ToInt(map1[name], dv)
}

func (r *Request) ParamInt64(name string, defaultValue ...interface{}) int64 {
	dv := int64(math.MinInt64)

	if len(defaultValue) > 0 {
		if n1, err := castx.ToInt64E(defaultValue[0]); err == nil {
			dv = n1
		}
	}

	map1 := map[string]string{}

	for name, value := range r.GetQueryParams() {
		map1[name] = value
	}

	for name, value := range r.GetFormData() {
		map1[name] = value
	}

	return castx.ToInt64(map1[name], dv)
}

func (r *Request) ParamFloat32(name string, defaultValue ...interface{}) float32 {
	dv := float32(math.SmallestNonzeroFloat32)

	if len(defaultValue) > 0 {
		if n1, err := castx.ToFloat32E(defaultValue[0]); err == nil {
			dv = n1
		}
	}

	map1 := map[string]string{}

	for name, value := range r.GetQueryParams() {
		map1[name] = value
	}

	for name, value := range r.GetFormData() {
		map1[name] = value
	}

	return castx.ToFloat32(map1[name], dv)
}

func (r *Request) ParamFloat64(name string, defaultValue ...interface{}) float64 {
	dv := math.SmallestNonzeroFloat64

	if len(defaultValue) > 0 {
		if n1, err := castx.ToFloat64E(defaultValue[0]); err == nil {
			dv = n1
		}
	}

	map1 := map[string]string{}

	for name, value := range r.GetQueryParams() {
		map1[name] = value
	}

	for name, value := range r.GetFormData() {
		map1[name] = value
	}

	return castx.ToFloat64(map1[name], dv)
}

func (r *Request) GetJwt() *jwt.Token {
	token := strings.TrimSpace(r.ctx.Get(fiber.HeaderAuthorization))
	token = stringx.RegexReplace(token, `[\x20\t]+`, " ")

	if strings.Contains(token, " ") {
		token = stringx.SubstringAfter(token, " ")
	}

	if token == "" {
		return nil
	}

	tk, _ := ParseJsonWebToken(token)
	return tk
}

func (r *Request) JwtClaimString(name string, defaultValue ...interface{}) string {
	var dv string

	if len(defaultValue) > 0 {
		if s1, err := castx.ToStringE(defaultValue[0]); err == nil {
			dv = s1
		}
	}

	token := strings.TrimSpace(r.ctx.Get(fiber.HeaderAuthorization))
	token = stringx.RegexReplace(token, `[\x20\t]+`, " ")

	if strings.Contains(token, " ") {
		token = stringx.SubstringAfter(token, " ")
	}

	if token == "" {
		return dv
	}

	return JwtClaimString(token, name, dv)
}

func (r *Request) JwtClaimBool(name string, defaultValue ...interface{}) bool {
	var dv bool

	if len(defaultValue) > 0 {
		if b1, err := castx.ToBoolE(defaultValue[0]); err == nil {
			dv = b1
		}
	}

	token := strings.TrimSpace(r.ctx.Get(fiber.HeaderAuthorization))
	token = stringx.RegexReplace(token, `[\x20\t]+`, " ")

	if strings.Contains(token, " ") {
		token = stringx.SubstringAfter(token, " ")
	}

	if token == "" {
		return dv
	}

	return JwtClaimBool(token, name, dv)
}

func (r *Request) JwtClaimInt(name string, defaultValue ...interface{}) int {
	dv := math.MinInt32

	if len(defaultValue) > 0 {
		if n1, err := castx.ToIntE(defaultValue[0]); err == nil {
			dv = n1
		}
	}

	token := strings.TrimSpace(r.ctx.Get(fiber.HeaderAuthorization))
	token = stringx.RegexReplace(token, `[\x20\t]+`, " ")

	if strings.Contains(token, " ") {
		token = stringx.SubstringAfter(token, " ")
	}

	if token == "" {
		return dv
	}

	return JwtClaimInt(token, name, dv)
}

func (r *Request) JwtClaimInt64(name string, defaultValue ...interface{}) int64 {
	dv := int64(math.MinInt64)

	if len(defaultValue) > 0 {
		if n1, err := castx.ToInt64E(defaultValue[0]); err == nil {
			dv = n1
		}
	}

	token := strings.TrimSpace(r.ctx.Get(fiber.HeaderAuthorization))
	token = stringx.RegexReplace(token, `[\x20\t]+`, " ")

	if strings.Contains(token, " ") {
		token = stringx.SubstringAfter(token, " ")
	}

	if token == "" {
		return dv
	}

	return JwtClaimInt64(token, name, dv)
}

func (r *Request) JwtClaimFloat32(name string, defaultValue ...interface{}) float32 {
	dv := float32(math.SmallestNonzeroFloat32)

	if len(defaultValue) > 0 {
		if n1, err := castx.ToFloat32E(defaultValue[0]); err == nil {
			dv = n1
		}
	}

	token := strings.TrimSpace(r.ctx.Get(fiber.HeaderAuthorization))
	token = stringx.RegexReplace(token, `[\x20\t]+`, " ")

	if strings.Contains(token, " ") {
		token = stringx.SubstringAfter(token, " ")
	}

	if token == "" {
		return dv
	}

	return JwtClaimFloat32(token, name, dv)
}

func (r *Request) JwtClaimFloat64(name string, defaultValue ...interface{}) float64 {
	dv := math.SmallestNonzeroFloat64

	if len(defaultValue) > 0 {
		if n1, err := castx.ToFloat64E(defaultValue[0]); err == nil {
			dv = n1
		}
	}

	token := strings.TrimSpace(r.ctx.Get(fiber.HeaderAuthorization))
	token = stringx.RegexReplace(token, `[\x20\t]+`, " ")

	if strings.Contains(token, " ") {
		token = stringx.SubstringAfter(token, " ")
	}

	if token == "" {
		return dv
	}

	return JwtClaimFloat64(token, name, dv)
}

func (r *Request) JwtClaimStringSlice(name string) []string {
	token := strings.TrimSpace(r.ctx.Get(fiber.HeaderAuthorization))
	token = stringx.RegexReplace(token, `[\x20\t]+`, " ")

	if strings.Contains(token, " ") {
		token = stringx.SubstringAfter(token, " ")
	}

	if token == "" {
		return make([]string, 0)
	}

	return JwtClaimStringSlice(token, name)
}

func (r *Request) JwtClaimIntSlice(name string) []int {
	token := strings.TrimSpace(r.ctx.Get(fiber.HeaderAuthorization))
	token = stringx.RegexReplace(token, `[\x20\t]+`, " ")

	if strings.Contains(token, " ") {
		token = stringx.SubstringAfter(token, " ")
	}

	if token == "" {
		return make([]int, 0)
	}

	return JwtClaimIntSlice(token, name)
}

func (r *Request) GetRawBody() []byte {
	method := r.ctx.Method()
	contentType := strings.ToLower(r.GetHeader(fiber.HeaderContentType))
	isPostForm := strings.Contains(contentType, fiber.MIMEApplicationForm)
	isMultipartForm := strings.Contains(contentType, fiber.MIMEMultipartForm)

	if method == "POST" && (isPostForm || isMultipartForm) {
		formData := r.GetFormData()

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

	rrb := r.ctx.Locals("requestRawBody")

	if buf, ok := rrb.([]byte); ok && len(buf) > 0 {
		if AppConf.GetBoolean("logging.logGetRawBody") {
			RuntimeLogger().Debug("raw body: " + string(buf))
		}

		return buf
	}

	return make([]byte, 0)
}

// @param string[]|string rules
func (r *Request) GetMap(rules ...interface{}) map[string]interface{} {
	method := r.ctx.Method()
	methods := []string{"POST", "PUT", "PATCH", "DELETE"}
	contentType := strings.ToLower(r.ctx.Get(fiber.HeaderContentType))
	isPostForm := strings.Contains(contentType, fiber.MIMEApplicationForm)
	isMultipartForm := strings.Contains(contentType, fiber.MIMEMultipartForm)
	isJson := strings.Contains(contentType, fiber.MIMEApplicationJSON)
	isXml1 := strings.Contains(contentType, fiber.MIMEApplicationXML)
	isXml2 := strings.Contains(contentType, fiber.MIMETextXML)
	map1 := map[string]interface{}{}

	if method == "GET" {
		for key, value := range r.GetQueryParams() {
			map1[key] = value
		}
	} else if method == "POST" && (isPostForm || isMultipartForm) {
		for key, value := range r.GetQueryParams() {
			map1[key] = value
		}

		for key, value := range r.GetFormData() {
			map1[key] = value
		}
	} else if slicex.InStringSlice(method, methods) {
		return map1
	} else if isJson {
		map1 = jsonx.MapFrom(r.GetRawBody())
	} else if isXml1 || isXml2 {
		map2 := mapx.FromXml(r.GetRawBody())

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

func (r *Request) DtoBind(dto interface{}) error {
	return mapx.BindToDto(r.GetMap(), dto)
}

func (r *Request) GetUploadedFile(formFieldName string) *multipart.FileHeader {
	if fh, err := r.ctx.FormFile(formFieldName); err != nil {
		return fh
	}

	return nil
}

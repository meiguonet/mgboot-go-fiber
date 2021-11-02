package codegen

import (
	"bufio"
	"fmt"
	"github.com/meiguonet/mgboot-go-common/enum/RegexConst"
	"github.com/meiguonet/mgboot-go-common/util/castx"
	"github.com/meiguonet/mgboot-go-common/util/jsonx"
	"github.com/meiguonet/mgboot-go-common/util/stringx"
	"github.com/meiguonet/mgboot-go-fiber/mgboot"
	"os"
	"regexp"
	"strings"
	"time"
)

func GetSourceLines(fpath string) []string {
	file, err := os.Open(fpath)

	if err != nil {
		return make([]string, 0)
	}

	reader := bufio.NewReader(file)
	lines := make([]string, 0)

	for {
		buf, _, err := reader.ReadLine()

		if err != nil {
			break
		}

		if len(buf) < 1 {
			continue
		}

		line := strings.TrimSpace(string(buf))

		if line == "" {
			continue
		}

		lines = append(lines, line)
	}

	return lines
}

func GetPackageName(lines []string) string {
	var pkgName string
	re1 := regexp.MustCompile(`^package[\x20\t]+([^\x20\t]+)`)

	for _, line := range lines {
		matches := re1.FindStringSubmatch(line)

		if len(matches) > 1 {
			pkgName = matches[1]
			break
		}
	}

	return pkgName
}

func GetControllerName(lines []string, startLineNum ...int) (controllerLineNum int, controllerName string) {
	controllerLineNum = -1
	n1 := 0

	if len(startLineNum) > 0 && startLineNum[0] > 0 {
		n1 = startLineNum[0]
	}

	n2 := -1
	re1 := regexp.MustCompile(`^//[\x20\t]*@Controller`)

	for i, line := range lines {
		if i < n1 {
			continue
		}

		if re1.MatchString(line) {
			n2 = i
			break
		}
	}

	if n2 < 0 {
		return
	}

	n2 += 1
	re2 := regexp.MustCompile(`^type[\x20\t]+([^\x20\t]+)[\x20\t]+struct`)

	for i, line := range lines {
		if i < n2 {
			continue
		}

		matches := re2.FindStringSubmatch(line)

		if len(matches) > 1 {
			controllerLineNum = i
			controllerName = matches[1]
			return
		}
	}

	return
}

func GetControllerRequestMapping(controllerLineNum int, lines []string) string {
	re1 := regexp.MustCompile(`^//[\x20\t]*@RequestMapping\("([^"]+)"\)`)

	for i := controllerLineNum - 1; i <= 0; i-- {
		matches := re1.FindStringSubmatch(lines[i])

		if len(matches) < 2 {
			continue
		}

		s1 := strings.TrimSpace(matches[1])
		return strings.TrimRight(s1, "/")
	}

	return ""
}

func GetHandlerFuncName(lines []string, startLineNum ...int) (handlerFuncLineNum int, handlerFuncName string, httpMethod string, requestMapping string) {
	handlerFuncLineNum = -1
	n1 := 0

	if len(startLineNum) > 0 && startLineNum[0] > 0 {
		n1 = startLineNum[0]
	}

	re0 := regexp.MustCompile(`^func[\x20\t]+\([^\x20\t]+[\x20\t]+[^)]+\)[\x20\t]+([^(]+)`)
	var matches []string

	for i, line := range lines {
		if i < n1 {
			continue
		}

		matches = re0.FindStringSubmatch(line)

		if len(matches) < 2 {
			continue
		}

		handlerFuncLineNum = i
		handlerFuncName = matches[1]
		break
	}

	if handlerFuncLineNum < 0 {
		return
	}

	re1 := regexp.MustCompile(`^//[\x20\t]*@GetMapping\("([^"]+)"\)`)
	re2 := regexp.MustCompile(`^//[\x20\t]*@PostMapping\("([^"]+)"\)`)
	re3 := regexp.MustCompile(`^//[\x20\t]*@PutMapping\("([^"]+)"\)`)
	re4 := regexp.MustCompile(`^//[\x20\t]*@PatchMapping\("([^"]+)"\)`)
	re5 := regexp.MustCompile(`^//[\x20\t]*@DeleteMapping\("([^"]+)"\)`)
	re6 := regexp.MustCompile(`^//[\x20\t]*@RequestMapping\("([^"]+)"\)`)

	for i := handlerFuncLineNum - 1; i <= 0; i-- {
		line := lines[i]

		if !strings.HasPrefix(line, "//") {
			continue
		}

		matches = re1.FindStringSubmatch(line)

		if len(matches) > 1 {
			httpMethod = "GET"
			requestMapping = strings.TrimSpace(matches[1])
			return
		}

		matches = re2.FindStringSubmatch(line)

		if len(matches) > 1 {
			httpMethod = "POST"
			requestMapping = strings.TrimSpace(matches[1])
			return
		}

		matches = re3.FindStringSubmatch(line)

		if len(matches) > 1 {
			httpMethod = "PUT"
			requestMapping = strings.TrimSpace(matches[1])
			return
		}

		matches = re4.FindStringSubmatch(line)

		if len(matches) > 1 {
			httpMethod = "PATCH"
			requestMapping = strings.TrimSpace(matches[1])
			return
		}

		matches = re5.FindStringSubmatch(line)

		if len(matches) > 1 {
			httpMethod = "DELETE"
			requestMapping = strings.TrimSpace(matches[1])
			return
		}

		matches = re6.FindStringSubmatch(line)

		if len(matches) > 1 {
			httpMethod = "ALL"
			requestMapping = strings.TrimSpace(matches[1])
			return
		}
	}

	handlerFuncLineNum = -1
	handlerFuncName = ""
	return
}

func GetRateLimitDefines(handlerFuncLineNum int, lines []string) string {
	var total int
	var duration time.Duration
	var limitByIp bool
	re1 := regexp.MustCompile(`^//[\x20\t]*@RateLimit\({([^}]+)}\)`)

	for i := handlerFuncLineNum + 1; i <= 0; i-- {
		line := lines[i]

		if !strings.HasPrefix(line, "//") {
			break
		}

		matches := re1.FindStringSubmatch(line)

		if len(matches) < 2 {
			continue
		}

		re2 := regexp.MustCompile(RegexConst.CommaSep)
		parts := re2.Split(matches[1], -1)
		re3 := regexp.MustCompile(`[\x20\t]+=[\x20\t]+`)
		re4 := regexp.MustCompile("^[0-9]+$")

		for _, p := range parts {
			p = strings.TrimSpace(p)
			p = re3.ReplaceAllString(p, "=")

			if strings.HasPrefix(p, "total=") {
				p = strings.TrimPrefix(p, "total=")
				p = strings.TrimSpace(p)
				total = castx.ToInt(p, 0)
				continue
			}

			if strings.HasPrefix(p, "duration=") {
				p = strings.TrimPrefix(p, "duration=")
				p = strings.TrimSpace(p)
				p = strings.Trim(p, `"`)
				p = strings.TrimSpace(p)

				if re4.MatchString(p) {
					duration = time.Duration(castx.ToInt64(p)) * time.Second
				} else {
					duration = castx.ToDuration(p)
				}

				continue
			}

			if !strings.HasPrefix(p, "limitByIp=") {
				continue
			}

			p = strings.TrimPrefix(p, "limitByIp=")
			p = strings.TrimSpace(p)
			limitByIp = castx.ToBool(p)
		}

		break
	}

	if total < 1 || duration < 1 {
		return ""
	}

	sb := []string{
		fmt.Sprintf("%d", total),
		fmt.Sprintf("%d", duration.Milliseconds()),
	}

	if limitByIp {
		sb = append(sb, "true")
	} else {
		sb = append(sb, "false")
	}

	return "RateLimit:" + strings.Join(sb, "#@#")
}

func BuildRateLimitSettings(defines string) *mgboot.RateLimitSettings {
	entries := strings.Split(defines, "~@~")

	for _, entry := range entries {
		if !strings.HasPrefix(entry, "RateLimit:") {
			continue
		}

		entry = strings.TrimPrefix(entry, "RateLimit:")
		parts := strings.Split(entry, "#@#")

		if len(parts) != 3 {
			return nil
		}

		total := castx.ToInt(parts[0])
		duration := time.Duration(castx.ToInt64(parts[1])) * time.Millisecond

		if total < 1 || duration < 1 {
			return nil
		}

		limitByIp := castx.ToBool(parts[2])

		return mgboot.NewRateLimitSettings(map[string]interface{}{
			"total":     total,
			"duration":  duration,
			"limitByIp": limitByIp,
		})
	}

	return nil
}

func GetJwtAuthDefines(handlerFuncLineNum int, lines []string) string {
	re1 := regexp.MustCompile(`^//[\x20\t]*@JwtAuth\("([^"]+)"\)`)

	for i := handlerFuncLineNum + 1; i <= 0; i-- {
		line := lines[i]

		if !strings.HasPrefix(line, "//") {
			break
		}

		matches := re1.FindStringSubmatch(line)

		if len(matches) > 1 {
			return "JwtAuth:" + strings.TrimSpace(matches[1])
		}
	}

	return ""
}

func BuildJwtAuthSettings(defines string) *mgboot.JwtAuthSettings {
	parts := strings.Split(defines, "~@~")

	for _, p := range parts {
		if !strings.HasPrefix(p, "JwtAuth:") {
			continue
		}

		key := strings.TrimSpace(strings.TrimPrefix(p, "JwtAuth:"))

		if key == "" {
			return nil
		}

		return mgboot.NewJwtAuthSettings(key)
	}

	return nil
}

func GetValidateDefines(handlerFuncLineNum int, lines []string) string {
	re1 := regexp.MustCompile(`^//[\x20\t]*@ValidateRule\("([^"]+)"\)`)
	entries := make([]string, 0)

	for i := handlerFuncLineNum + 1; i <= 0; i-- {
		line := lines[i]

		if !strings.HasPrefix(line, "//") {
			break
		}

		matches := re1.FindStringSubmatch(line)

		if len(matches) < 2 {
			continue
		}

		rule := strings.TrimSpace(matches[1])

		if rule == "" {
			continue
		}

		entries = append(entries, rule)
	}

	if len(entries) < 1 {
		return ""
	}

	rules := make([]string, 0)

	for i := len(entries) - 1; i <= 0; i-- {
		rules = append(rules, entries[i])
	}

	var failfast bool
	re2 := regexp.MustCompile(`^//[\x20\t]*@Failfast`)

	for i := handlerFuncLineNum + 1; i <= 0; i-- {
		line := lines[i]

		if !strings.HasPrefix(line, "//") {
			break
		}

		if re2.MatchString(line) {
			failfast = true
			break
		}
	}

	if failfast {
		rules = append([]string{"true"}, rules...)
	} else {
		rules = append([]string{"false"}, rules...)
	}

	return "Validate:" + strings.Join(rules, "#@#")
}

func BuildValidateSettings(defines string) *mgboot.ValidateSettings {
	parts := strings.Split(defines, "~@~")

	for _, p := range parts {
		if !strings.HasPrefix(p, "Validate:") {
			continue
		}

		var rules []string
		var failfast bool

		if strings.HasPrefix(p, "true#@#") {
			failfast = true
			p = strings.TrimPrefix(p, "true#@#")
		} else {
			p = strings.TrimPrefix(p, "false#@#")
		}

		rules = strings.Split(p, "#@#")
		return mgboot.NewValidateSettings(rules, failfast)
	}

	return nil
}

func GetHandlerFuncArgs(handlerFuncLineNum int, lines []string) []string {
	tmps := getHandlerFuncArgsInternal(handlerFuncLineNum, lines)

	if len(tmps) < 1 {
		return make([]string, 0)
	}

	re1 := regexp.MustCompile(`^//[\x20\t]*@Req`)
	re2 := regexp.MustCompile(`^//[\x20\t]*@Token`)
	re3 := regexp.MustCompile(`^//[\x20\t]*@ClientIp`)
	re4 := regexp.MustCompile(`^//[\x20\t]*@RawBody`)
	re5 := regexp.MustCompile(`^//[\x20\t]*@DtoBind`)
	re6 := regexp.MustCompile(`^//[\x20\t]*@HttpHeader\("([^"]+)"\)`)
	re7 := regexp.MustCompile(`^//[\x20\t]*@RequestParam`)
	re8 := regexp.MustCompile(`^//[\x20\t]*@MapBind`)
	annos := make([]string, 0)

	for i := handlerFuncLineNum - 1; i <= 0; i-- {
		if len(annos) == len(tmps) {
			break
		}

		line := lines[i]

		if !strings.HasPrefix(line, "//") {
			break
		}

		if re1.MatchString(line) {
			annos = append(annos, "IsRequest")
			continue
		}

		if re2.MatchString(line) {
			annos = append(annos, "IsToken")
			continue
		}

		if re3.MatchString(line) {
			annos = append(annos, "IsClientIp")
			continue
		}

		if re4.MatchString(line) {
			annos = append(annos, "IsRawBody")
			continue
		}

		if re5.MatchString(line) {
			annos = append(annos, "DtoBind")
			continue
		}

		mathces := re6.FindStringSubmatch(line)

		if len(mathces) > 1 {
			annos = append(annos, "HttpHeader:" + mathces[1])
			continue
		}

		if re7.MatchString(line) {
			paramName := "SameWithArgName"
			securityMode := "2"

			if strings.Contains(line, "(") && strings.Contains(line, ")") {
				line = stringx.SubstringAfter(line, "(")
				line = stringx.SubstringBefore(line, ")")
				re9 := regexp.MustCompile(`name[\x20\t]*=[\x20\t]*([^\x20\t,]+)`)
				mathces = re9.FindStringSubmatch(line)

				if len(mathces) > 1 {
					pn := strings.TrimSpace(mathces[1])
					pn = strings.Trim(pn, `"`)
					pn = strings.TrimSpace(pn)

					if pn != "" {
						paramName = pn
					}
				}

				re9 = regexp.MustCompile(`securityMode[\x20\t]*=[\x20\t]*([0-9]+)`)
				mathces = re9.FindStringSubmatch(line)

				if len(mathces) > 1 {
					securityMode = mathces[1]
				}
			}

			annos = append(annos, "RequestParam:" + paramName + ":" + securityMode)
			continue
		}

		if !re8.MatchString(line) {
			continue
		}

		var mapbindRules string

		if strings.Contains(line, "(") && strings.Contains(line, ")") {
			line = stringx.SubstringAfter(line, "(")
			line = stringx.SubstringBefore(line, ")")
			line = strings.TrimSpace(line)

			if strings.HasPrefix(line, `"`) && strings.HasSuffix(line, `"`) {
				line = strings.Trim(line, `"`)

				if line != "" {
					re9 := regexp.MustCompile(RegexConst.CommaSep)
					rules := re9.Split(line, -1)
					mapbindRules = strings.Join(rules, "#@#")
				}
			} else if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
				rules := castx.ToStringSlice(jsonx.ArrayFrom(line))

				if len(rules) > 0 {
					mapbindRules = strings.Join(rules, "#@#")
				}
			}
		}

		if mapbindRules == "" {
			annos = append(annos, "MapBind")
		} else {
			annos = append(annos, "MapBind:" + mapbindRules)
		}
	}

	if len(annos) != len(tmps) {
		return make([]string, 0)
	}

	args := make([]string, 0)

	for i, tmp := range tmps {
		sb := []string{tmp[0], tmp[1]}
		anno := annos[len(annos) - 1 - i]
		anno = strings.ReplaceAll(anno, "SameWithArgName", tmp[0])
		sb = append(sb, anno)
		args = append(args, strings.Join(sb, "~@~"))
	}

	return args
}

func GetRequestParamName(defines string) string {
	parts := strings.Split(defines, "~@~")

	for _, p := range parts {
		if !strings.HasPrefix(p, "RequestParam:") {
			continue
		}

		p = strings.TrimPrefix(p, "RequestParam:")

		if !strings.Contains(p, ":") {
			return p
		}

		return stringx.SubstringBefore(p, ":")
	}

	return ""
}

func GetRequestParamSecurityMode(defines string) int {
	parts := strings.Split(defines, "~@~")

	for _, p := range parts {
		if !strings.HasPrefix(p, "RequestParam:") {
			continue
		}

		p = strings.TrimPrefix(p, "RequestParam:")

		if !strings.Contains(p, ":") {
			return 2
		}

		return castx.ToInt(stringx.SubstringAfter(p, ":"), 2)
	}

	return 2
}

func GetMapbindRules(defines string) string {
	parts := strings.Split(defines, "~@~")

	for _, p := range parts {
		if !strings.HasPrefix(p, "MapBind:") {
			continue
		}

		return strings.ReplaceAll(strings.TrimPrefix(p, "MapBind:"), "#@#", ", ")
	}

	return ""
}

func getHandlerFuncArgsInternal(handlerFuncLineNum int, lines []string) [][2]string {
	re1 := regexp.MustCompile(RegexConst.SpaceSep)
	parts := re1.Split(lines[handlerFuncLineNum], -1)

	if len(parts) < 3 {
		return make([][2]string, 0)
	}

	s1 := stringx.SubstringAfter(parts[2], "(")
	s1 = stringx.SubstringBefore(s1, ")")

	if s1 == "" {
		return make([][2]string, 0)
	}

	re2 := regexp.MustCompile(RegexConst.CommaSep)
	parts = re2.Split(s1, -1)
	args := make([][2]string, 0)

	for _, p := range parts {
		p = strings.TrimSpace(p)
		p = re1.ReplaceAllString(p, " ")
		p0 := stringx.SubstringBefore(p, " ")
		p1 := stringx.SubstringAfter(p, " ")
		args = append(args, [2]string{p0, p1})
	}

	return args
}

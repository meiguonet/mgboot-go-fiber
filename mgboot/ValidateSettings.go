package mgboot

import (
	"github.com/meiguonet/mgboot-go-common/enum/RegexConst"
	"github.com/meiguonet/mgboot-go-common/util/castx"
	"github.com/meiguonet/mgboot-go-common/util/stringx"
	"strings"
)

type ValidateSettings struct {
	rules    []string
	failfast bool
}

func NewValidateSettings(settings interface{}) *ValidateSettings {
	if map1, ok := settings.(map[string]interface{}); ok && len(map1) > 0 {
		return newValidateSettingsFromMap(map1)
	}

	if s1, ok := settings.(string); ok && s1 != "" {
		return newValidateSettingsFromString(s1)
	}

	return &ValidateSettings{rules: make([]string, 0)}
}

func newValidateSettingsFromMap(settings map[string]interface{}) *ValidateSettings {
	rules := make([]string, 0)

	if a1, ok := settings["rules"].([]string); ok && len(a1) > 0 {
		rules = a1
	} else if s1, ok := settings["rules"].(string); ok && s1 != "" {
		rules = stringx.SplitWithRegexp(s1, RegexConst.CommaSep)
	}

	var failfast bool

	if b1, ok := settings["failfast"].(bool); ok {
		failfast = b1
	}

	return &ValidateSettings{rules: rules, failfast: failfast}
}

func newValidateSettingsFromString(defines string) *ValidateSettings {
	if strings.Contains(defines, "Validate:") {
		defines = strings.TrimPrefix(defines, "Validate:")
	}

	defines = strings.ReplaceAll(defines, "#@#", ",")
	parts := strings.Split(defines, "~@~")
	rules := stringx.SplitWithRegexp(parts[0], RegexConst.CommaSep)
	var failfast bool

	if len(parts) > 1 {
		if b1, err := castx.ToBoolE(parts[1]); err == nil {
			failfast = b1
		}
	}

	return &ValidateSettings{rules: rules, failfast: failfast}
}

func (st *ValidateSettings) Rules() []string {
	return st.rules
}

func (st *ValidateSettings) Failfast() bool {
	return st.failfast
}

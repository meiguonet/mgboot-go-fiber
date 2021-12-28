package mgboot

import (
	"github.com/meiguonet/mgboot-go-common/enum/RegexConst"
	"github.com/meiguonet/mgboot-go-common/util/jsonx"
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
	defines = strings.ReplaceAll(defines, "[syh]", `"`)
	parts := jsonx.ArrayFrom(defines)
	rules := make([]string, 0)
	var failfast bool

	for _, p := range parts {
		s1, ok := p.(string)

		if !ok || s1 == "" {
			continue
		}

		if s1 == "false" {
			continue
		}

		if s1 == "true" {
			failfast = true
			continue
		}

		rules = append(rules, s1)
	}

	return &ValidateSettings{rules: rules, failfast: failfast}
}

func (st *ValidateSettings) Rules() []string {
	return st.rules
}

func (st *ValidateSettings) Failfast() bool {
	return st.failfast
}

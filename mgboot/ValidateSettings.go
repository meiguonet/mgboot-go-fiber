package mgboot

import (
	"github.com/meiguonet/mgboot-go-common/enum/RegexConst"
	"regexp"
)

type ValidateSettings struct {
	rules    []string
	failfast bool
}

// @param string[]|string rules
func NewValidateSettings(rules interface{}, failfast bool) *ValidateSettings {
	var _rules []string

	if a1, ok := rules.([]string); ok {
		_rules = a1
	} else if s1, ok := rules.(string); ok && s1 != "" {
		re := regexp.MustCompile(RegexConst.CommaSep)
		_rules = re.Split(s1, -1)
	}

	return &ValidateSettings{rules: _rules, failfast: failfast}
}

func (st *ValidateSettings) Rules() []string {
	return st.rules
}

func (st *ValidateSettings) Failfast() bool {
	return st.failfast
}

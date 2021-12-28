package mgboot

import (
	"github.com/meiguonet/mgboot-go-common/util/castx"
	"github.com/meiguonet/mgboot-go-common/util/jsonx"
	"strings"
	"time"
)

type RateLimitSettings struct {
	total     int
	duration  time.Duration
	limitByIp bool
}

func NewRateLimitSettings(settings interface{}) *RateLimitSettings {
	if map1, ok := settings.(map[string]interface{}); ok && len(map1) > 0 {
		return newRateLimitSettingsFromMap(map1)
	}

	if s1, ok := settings.(string); ok && s1 != "" {
		return newRateLimitSettingsFromString(s1)
	}

	return &RateLimitSettings{}
}

func newRateLimitSettingsFromMap(settings map[string]interface{}) *RateLimitSettings {
	var total int

	if n1, ok := settings["total"].(int); ok {
		total = n1
	}

	var duration time.Duration

	if d1, ok := settings["duration"].(time.Duration); ok {
		duration = d1
	} else if s1, ok := settings["duration"].(string); ok && s1 != "" {
		duration = castx.ToDuration(s1)
	}

	var limitByIp bool

	if b1, ok := settings["limitByIp"].(bool); ok {
		limitByIp = b1
	}

	return &RateLimitSettings{
		total:     total,
		duration:  duration,
		limitByIp: limitByIp,
	}
}

func newRateLimitSettingsFromString(defines string) *RateLimitSettings {
	defines = strings.ReplaceAll(defines, "[syh]", `"`)
	map1 := jsonx.MapFrom(defines)
	total := castx.ToInt(map1["total"], 0)
	var duration time.Duration
	n1 := castx.ToInt64(map1["duration"])

	if n1 > 0 {
		duration = time.Duration(n1) * time.Millisecond
	}

	return &RateLimitSettings{
		total:     total,
		duration:  duration,
		limitByIp: castx.ToBool(map1["limitByIp"]),
	}
}

func (st *RateLimitSettings) Total() int {
	return st.total
}

func (st *RateLimitSettings) Duration() time.Duration {
	return st.duration
}

func (st *RateLimitSettings) LimitByIp() bool {
	return st.limitByIp
}

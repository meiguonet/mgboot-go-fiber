package mgboot

import (
	"github.com/meiguonet/mgboot-go-common/util/castx"
	"time"
)

type RateLimitSettings struct {
	total     int
	duration  time.Duration
	limitByIp bool
}

func NewRateLimitSettings(settings map[string]interface{}) *RateLimitSettings {
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

func (st *RateLimitSettings) Total() int {
	return st.total
}

func (st *RateLimitSettings) Duration() time.Duration {
	return st.duration
}

func (st *RateLimitSettings) LimitByIp() bool {
	return st.limitByIp
}

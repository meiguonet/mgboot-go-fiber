package mgboot

import "strings"

type JwtAuthSettings struct {
	key string
}

func NewJwtAuthSettings(key string) *JwtAuthSettings {
	if strings.Contains(key, "JwtAuth:") {
		key = strings.TrimPrefix(key, "JwtAuth:")
	}

	return &JwtAuthSettings{key: key}
}

func (st *JwtAuthSettings) Key() string {
	return st.key
}

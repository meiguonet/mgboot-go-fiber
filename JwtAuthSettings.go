package mgboot

type JwtAuthSettings struct {
	key string
}

func NewJwtAuthSettings(key string) *JwtAuthSettings {
	return &JwtAuthSettings{key: key}
}

func (st *JwtAuthSettings) Key() string {
	return st.key
}

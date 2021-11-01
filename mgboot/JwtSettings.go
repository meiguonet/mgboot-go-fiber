package mgboot

import (
	"github.com/meiguonet/mgboot-go-common/util/castx"
	"time"
)

type JwtSettings struct {
	issuer            string
	ttl               time.Duration
	refreshTokenTtl   time.Duration
	publicKeyPemFile  string
	privateKeyPemFile string
}

func NewJwtSettings(settings map[string]interface{}) *JwtSettings {
	var issuer string

	if s1, ok := settings["issuer"].(string); ok && s1 != "" {
		issuer = s1
	} else if s1, ok := settings["iss"].(string); ok && s1 != "" {
		issuer = s1
	}

	var ttl time.Duration

	if d1, ok := settings["ttl"].(time.Duration); ok {
		ttl = d1
	} else if s1, ok := settings["ttl"].(string); ok && s1 != "" {
		ttl = castx.ToDuration(ttl)
	}

	var refreshTokenTtl time.Duration

	if d1, ok := settings["refreshTokenTtl"].(time.Duration); ok {
		refreshTokenTtl = d1
	} else if s1, ok := settings["refreshTokenTtl"].(string); ok && s1 != "" {
		refreshTokenTtl = castx.ToDuration(ttl)
	}

	return &JwtSettings{
		issuer:            issuer,
		ttl:               ttl,
		refreshTokenTtl:   refreshTokenTtl,
		publicKeyPemFile:  castx.ToString(settings["publicKeyPemFile"]),
		privateKeyPemFile: castx.ToString(settings["privateKeyPemFile"]),
	}
}

func (st *JwtSettings) Issuer() string {
	return st.issuer
}

func (st *JwtSettings) Ttl() time.Duration {
	return st.ttl
}

func (st *JwtSettings) RefreshTokenTtl() time.Duration {
	return st.refreshTokenTtl
}

func (st *JwtSettings) PublicKeyPemFile() string {
	return st.publicKeyPemFile
}

func (st *JwtSettings) PrivateKeyPemFile() string {
	return st.privateKeyPemFile
}

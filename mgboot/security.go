package mgboot

import (
	"crypto/rsa"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-errors/errors"
	"github.com/gofiber/fiber/v2"
	"github.com/meiguonet/mgboot-go-common/AppConf"
	"github.com/meiguonet/mgboot-go-common/util/castx"
	"github.com/meiguonet/mgboot-go-common/util/fsx"
	"github.com/meiguonet/mgboot-go-fiber/enum/JwtVerifyErrno"
	"io/ioutil"
	"math"
	"os"
	"time"
)

var corsSettings *CorsSettings
var jwtPublicKeyPemFile string
var jwtPrivateKeyPemFile string
var jwtSettings map[string]*JwtSettings

func WithCorsSettings(settings ...map[string]interface{}) {
	_settings := map[string]interface{}{}

	if len(settings) > 0 && len(settings[0]) > 0 {
		_settings = settings[0]
	}

	if len(_settings) < 1 {
		_settings = AppConf.GetMap("cors")
	}
}

func GetCorsSettings() *CorsSettings {
	return corsSettings
}

func WithJwtPublicKeyPemFile(fpath string) {
	fpath = fsx.GetRealpath(fpath)

	if stat, err := os.Stat(fpath); err == nil && !stat.IsDir() {
		jwtPublicKeyPemFile = fpath
	}
}

func GetJwtPublicKeyPemFile() string {
	return jwtPublicKeyPemFile
}

func WithJwtPrivateKeyPemFile(fpath string) {
	fpath = fsx.GetRealpath(fpath)

	if stat, err := os.Stat(fpath); err == nil && !stat.IsDir() {
		jwtPrivateKeyPemFile = fpath
	}
}

func GetJwtPrivateKeyPemFile() string {
	return jwtPrivateKeyPemFile
}

func WithJwtSettings(key string, settings ...map[string]interface{}) {
	_settings := map[string]interface{}{}

	if len(settings) > 0 && len(settings[0]) > 0 {
		_settings = settings[0]
	}

	if len(_settings) < 1 {
		_settings = AppConf.GetMap("jwt." + key)
	}

	if _, ok := _settings["publicKeyPemFile"]; !ok {
		_settings["publicKeyPemFile"] = jwtPublicKeyPemFile
	}

	if _, ok := _settings["privateKeyPemFile"]; !ok {
		_settings["privateKeyPemFile"] = jwtPrivateKeyPemFile
	}

	if len(jwtSettings) < 1 {
		jwtSettings = map[string]*JwtSettings{key: NewJwtSettings(_settings)}
	} else {
		jwtSettings[key] = NewJwtSettings(_settings)
	}
}

func GetJwtSettings(key string) *JwtSettings {
	if len(jwtSettings) < 1 {
		return nil
	}

	return jwtSettings[key]
}

func ParseJsonWebToken(token string, pubpem ...string) (*jwt.Token, error) {
	var fpath string

	if len(pubpem) > 0 && pubpem[0] != "" {
		fpath = pubpem[0]
	}

	keyBytes := loadKeyPem("pub", fpath)

	if len(keyBytes) < 1 {
		return nil, errors.New("fail to load public key from pem file")
	}

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(keyBytes)

	if err != nil {
		return nil, err
	}

	return jwt.Parse(token, func(tk *jwt.Token) (interface{}, error) {
		if _, ok := tk.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", tk.Header["alg"])
		}

		return publicKey, nil
	})
}

// @param *jwt.Token|string arg0
func VerifyJsonWebToken(arg0 interface{}, settings *JwtSettings) int {
	var token *jwt.Token

	if tk, ok := arg0.(*jwt.Token); ok {
		token = tk
	} else if s1, ok := arg0.(string); ok && s1 != "" {
		tk, _ := ParseJsonWebToken(s1, settings.publicKeyPemFile)
		token = tk
	}

	if token == nil || !token.Valid {
		return JwtVerifyErrno.Invalid
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok {
		return JwtVerifyErrno.Invalid
	}

	iss := settings.Issuer()

	if iss != "" && castx.ToString(claims["iss"]) != iss {
		return JwtVerifyErrno.Invalid
	}

	exp := castx.ToInt64(claims["exp"])

	if exp > 0 && time.Now().Unix() > exp {
		return JwtVerifyErrno.Expired
	}

	return 0
}

// @param *JwtSettings|string arg0
func BuildJsonWebToken(arg0 interface{}, isRefreshToken bool, claims ...map[string]interface{}) (token string, err error) {
	var settings *JwtSettings

	if s1, ok := arg0.(*JwtSettings); ok && s1 != nil {
		settings = s1
	} else if s1, ok := arg0.(string); ok && s1 != "" {
		settings = GetJwtSettings(s1)
	}

	if settings == nil {
		err = errors.New("in mgboot.BuildJsonWebToken function, *JwtSettings is nil")
		return
	}

	keyBytes := loadKeyPem("pri", settings.privateKeyPemFile)

	if len(keyBytes) < 1 {
		err = errors.New("in mgboot.BuildJsonWebToken function, fail to load private key from pem file")
		return
	}

	var privateKey *rsa.PrivateKey
	privateKey, err = jwt.ParseRSAPrivateKeyFromPEM(keyBytes)

	if err != nil {
		return
	}

	var exp int64

	if isRefreshToken {
		exp = time.Now().Add(settings.RefreshTokenTtl()).Unix()
	} else {
		exp = time.Now().Add(settings.Ttl()).Unix()
	}

	mapClaims := jwt.MapClaims{
		"iss": settings.Issuer(),
		"exp": exp,
	}

	if len(claims) > 0 && len(claims[0]) > 0 {
		for claimName, claimValue := range claims[0] {
			mapClaims[claimName] = claimValue
		}
	}

	token, err = jwt.NewWithClaims(jwt.SigningMethodRS256, mapClaims).SignedString(privateKey)
	return
}

// @param *jwt.Token|*fiber.Ctx|string arg0
func JwtClaim(arg0 interface{}, name string, defaultValue ...interface{}) string {
	var dv string

	if len(defaultValue) > 0 {
		if s1, err := castx.ToStringE(defaultValue[0]); err == nil {
			dv = s1
		}
	}

	var token *jwt.Token

	if tk, ok := arg0.(*jwt.Token); ok {
		token = tk
	} else if ctx, ok := arg0.(*fiber.Ctx); ok {
		token = GetJwt(ctx)
	} else if s1, ok := arg0.(string); ok && s1 != "" {
		tk, _ := ParseJsonWebToken(s1)
		token = tk
	}

	if token == nil {
		return dv
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok {
		return dv
	}

	if s1 := castx.ToString(claims[name]); s1 != "" {
		return s1
	}

	return dv
}

// @param *jwt.Token|*fiber.Ctx|string arg0
func JwtClaimBool(arg0 interface{}, name string, defaultValue ...interface{}) bool {
	var dv bool

	if len(defaultValue) > 0 {
		if b1, err := castx.ToBoolE(defaultValue[0]); err == nil {
			dv = b1
		}
	}

	var token *jwt.Token

	if tk, ok := arg0.(*jwt.Token); ok {
		token = tk
	} else if ctx, ok := arg0.(*fiber.Ctx); ok {
		token = GetJwt(ctx)
	} else if s1, ok := arg0.(string); ok && s1 != "" {
		tk, _ := ParseJsonWebToken(s1)
		token = tk
	}

	if token == nil {
		return dv
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok {
		return dv
	}

	if b1, err := castx.ToBoolE(claims[name]); err == nil {
		return b1
	}

	return dv
}

// @param *jwt.Token|*fiber.Ctx|string arg0
func JwtClaimInt(arg0 interface{}, name string, defaultValue ...interface{}) int {
	dv := math.MinInt32

	if len(defaultValue) > 0 {
		if n1, err := castx.ToIntE(defaultValue[0]); err == nil {
			dv = n1
		}
	}

	var token *jwt.Token

	if tk, ok := arg0.(*jwt.Token); ok {
		token = tk
	} else if ctx, ok := arg0.(*fiber.Ctx); ok {
		token = GetJwt(ctx)
	} else if s1, ok := arg0.(string); ok && s1 != "" {
		tk, _ := ParseJsonWebToken(s1)
		token = tk
	}

	if token == nil {
		return dv
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok {
		return dv
	}

	if n1, err := castx.ToIntE(claims[name]); err == nil {
		return n1
	}

	return dv
}

// @param *jwt.Token|*fiber.Ctx|string arg0
func JwtClaimInt64(arg0 interface{}, name string, defaultValue ...interface{}) int64 {
	dv := int64(math.MinInt64)

	if len(defaultValue) > 0 {
		if n1, err := castx.ToInt64E(defaultValue[0]); err == nil {
			dv = n1
		}
	}

	var token *jwt.Token

	if tk, ok := arg0.(*jwt.Token); ok {
		token = tk
	} else if ctx, ok := arg0.(*fiber.Ctx); ok {
		token = GetJwt(ctx)
	} else if s1, ok := arg0.(string); ok && s1 != "" {
		tk, _ := ParseJsonWebToken(s1)
		token = tk
	}

	if token == nil {
		return dv
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok {
		return dv
	}

	if n1, err := castx.ToInt64E(claims[name]); err == nil {
		return n1
	}

	return dv
}

// @param *jwt.Token|*fiber.Ctx|string arg0
func JwtClaimFloat32(arg0 interface{}, name string, defaultValue ...interface{}) float32 {
	dv := float32(math.SmallestNonzeroFloat32)

	if len(defaultValue) > 0 {
		if n1, err := castx.ToFloat32E(defaultValue[0]); err == nil {
			dv = n1
		}
	}

	var token *jwt.Token

	if tk, ok := arg0.(*jwt.Token); ok {
		token = tk
	} else if ctx, ok := arg0.(*fiber.Ctx); ok {
		token = GetJwt(ctx)
	} else if s1, ok := arg0.(string); ok && s1 != "" {
		tk, _ := ParseJsonWebToken(s1)
		token = tk
	}

	if token == nil {
		return dv
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok {
		return dv
	}

	if n1, err := castx.ToFloat32E(claims[name]); err == nil {
		return n1
	}

	return dv
}

// @param *jwt.Token|*fiber.Ctx|string arg0
func JwtClaimFloat64(arg0 interface{}, name string, defaultValue ...interface{}) float64 {
	dv := math.SmallestNonzeroFloat64

	if len(defaultValue) > 0 {
		if n1, err := castx.ToFloat64E(defaultValue[0]); err == nil {
			dv = n1
		}
	}

	var token *jwt.Token

	if tk, ok := arg0.(*jwt.Token); ok {
		token = tk
	} else if ctx, ok := arg0.(*fiber.Ctx); ok {
		token = GetJwt(ctx)
	} else if s1, ok := arg0.(string); ok && s1 != "" {
		tk, _ := ParseJsonWebToken(s1)
		token = tk
	}

	if token == nil {
		return dv
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok {
		return dv
	}

	if n1, err := castx.ToFloat64E(claims[name]); err == nil {
		return n1
	}

	return dv
}

// @param *jwt.Token|*fiber.Ctx|string arg0
func JwtClaimStringSlice(arg0 interface{}, name string) []string {
	var token *jwt.Token

	if tk, ok := arg0.(*jwt.Token); ok {
		token = tk
	} else if ctx, ok := arg0.(*fiber.Ctx); ok {
		token = GetJwt(ctx)
	} else if s1, ok := arg0.(string); ok && s1 != "" {
		tk, _ := ParseJsonWebToken(s1)
		token = tk
	}

	if token == nil {
		return make([]string, 0)
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok {
		return make([]string, 0)
	}

	return castx.ToStringSlice(claims[name])
}

// @param *jwt.Token|*fiber.Ctx|string arg0
func JwtClaimIntSlice(arg0 interface{}, name string) []int {
	var token *jwt.Token

	if tk, ok := arg0.(*jwt.Token); ok {
		token = tk
	} else if ctx, ok := arg0.(*fiber.Ctx); ok {
		token = GetJwt(ctx)
	} else if s1, ok := arg0.(string); ok && s1 != "" {
		tk, _ := ParseJsonWebToken(s1)
		token = tk
	}

	if token == nil {
		return make([]int, 0)
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok {
		return make([]int, 0)
	}

	return castx.ToIntSlice(claims[name])
}

func loadKeyPem(typ string, arg1 interface{}) []byte {
	var fpath string

	if s1, ok := arg1.(string); ok && s1 != "" {
		fpath = s1
	} else if s1, ok := arg1.(*JwtSettings); ok && s1 != nil {
		switch typ {
		case "pub":
			fpath = s1.PublicKeyPemFile()
		case "pri":
			fpath = s1.PrivateKeyPemFile()
		}
	}

	if fpath == "" {
		switch typ {
		case "pub":
			fpath = GetJwtPublicKeyPemFile()
		case "pri":
			fpath = GetJwtPrivateKeyPemFile()
		}
	}

	if fpath == "" {
		return make([]byte, 0)
	}

	buf, err := ioutil.ReadFile(fpath)

	if err != nil {
		return make([]byte, 0)
	}

	return buf
}

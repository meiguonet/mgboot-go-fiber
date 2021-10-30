package mgboot

import (
	"github.com/meiguonet/mgboot-go-common/util/fsx"
	"os"
)

var ratelimiterLuaFile string
var ratelimiterCacheDir string

func WithRatelimiterLuaFile(fpath string) {
	fpath = fsx.GetRealpath(fpath)

	if stat, err := os.Stat(fpath); err == nil && !stat.IsDir() {
		ratelimiterLuaFile = fpath
	}
}

func RatelimiterLuaFile() string {
	return ratelimiterLuaFile
}

func WithRatelimiterCacheDir(dir string) {
	dir = fsx.GetRealpath(dir)

	if stat, err := os.Stat(dir); err == nil && stat.IsDir() {
		ratelimiterCacheDir = dir
	}
}

func RatelimiterCacheDir() string {
	return ratelimiterCacheDir
}

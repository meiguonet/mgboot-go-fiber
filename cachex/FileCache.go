package cachex

import (
	"fmt"
	"github.com/meiguonet/mgboot-go-common/util/castx"
	"github.com/meiguonet/mgboot-go-common/util/jsonx"
	"github.com/meiguonet/mgboot-go-common/util/securityx"
	"github.com/meiguonet/mgboot-go-common/util/stringx"
	"github.com/vmihailenco/msgpack/v5"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type CacheEntry struct {
	Data     string
	ExpireAt time.Time
}

type fileCache struct {
	mu sync.RWMutex
}

func newFileCache() *fileCache {
	return &fileCache{mu: sync.RWMutex{}}
}

func (c *fileCache) Get(key string, defaultValue ...interface{}) interface{} {
	var dv interface{}

	if len(defaultValue) > 0 {
		dv = defaultValue[0]
	}

	cacheKey := BuildCacheKey(key)
	c.mu.RLock()
	defer c.mu.RUnlock()
	dir := c.ensureCacheDirExists(cacheKey)

	if dir == "" {
		return dv
	}

	dir = strings.ReplaceAll(dir, "\\", "/")
	dir = strings.TrimRight(dir, "/")
	fpath := fmt.Sprintf("%s/%s.dat", dir, strings.ToLower(securityx.Md5(cacheKey)))
	buf, err := ioutil.ReadFile(fpath)

	if err != nil || len(buf) < 1 {
		return nil
	}

	var _entry CacheEntry

	if err := msgpack.Unmarshal(buf, &_entry); err != nil {
		return dv
	}

	if !_entry.ExpireAt.IsZero() && time.Now().Unix() > _entry.ExpireAt.Unix() {
		c.Delete(key)
		return dv
	}

	data := jsonx.MapFrom(_entry.Data)

	if value, ok := data["data"]; ok && value != nil {
		return value
	}

	return dv
}

func (c *fileCache) Set(key string, value interface{}, ttl ...interface{}) bool {
	cacheKey := BuildCacheKey(key)
	c.mu.Lock()
	defer c.mu.Unlock()
	dir := c.ensureCacheDirExists(cacheKey)

	if dir == "" {
		return false
	}

	dir = strings.ReplaceAll(dir, "\\", "/")
	dir = strings.TrimRight(dir, "/")
	var _ttl time.Duration

	if len(ttl) > 0 {
		switch t := ttl[0].(type) {
		case time.Duration:
			_ttl = t
		case int64:
			_ttl = time.Duration(t) * time.Second
		case int:
			_ttl = time.Duration(t) * time.Second
		case string:
			_ttl = castx.ToDuration(t)
		}
	}

	data := jsonx.ToJson(map[string]interface{}{"data": value})
	entry := CacheEntry{Data: data}

	if _ttl > 0 {
		entry.ExpireAt = time.Now().Add(_ttl)
	}

	buf, err := msgpack.Marshal(&entry)

	if err != nil || len(buf) < 1 {
		return false
	}

	fpath := fmt.Sprintf("%s/%s.dat", dir, strings.ToLower(securityx.Md5(cacheKey)))
	ioutil.WriteFile(fpath, buf, 0755)

	if stat, err := os.Stat(fpath); err == nil && !stat.IsDir() {
		return true
	}

	return false
}

func (c *fileCache) Delete(key string) bool {
	cacheKey := BuildCacheKey(key)
	dir := c.ensureCacheDirExists(cacheKey)

	if dir == "" {
		return true
	}

	dir = strings.ReplaceAll(dir, "\\", "/")
	dir = strings.TrimRight(dir, "/")
	fpath := fmt.Sprintf("%s/%s.dat", dir, strings.ToLower(securityx.Md5(cacheKey)))
	os.Remove(fpath)
	n1 := 0

	filepath.Walk(dir, func(spath string, _ os.FileInfo, err error) error {
		if err == nil {
			return nil
		}

		if spath == "." || spath == ".." {
			return nil
		}

		spath = strings.ReplaceAll(spath, "\\", "/")
		spath = strings.TrimRight(spath, "/")
		spath = stringx.EnsureLeft(spath, dir + "/")

		if spath != dir {
			n1++
		}

		return nil
	})

	if n1 > 0 {
		return true
	}

	os.Remove(dir)
	dir = stringx.SubstringBeforeLast(dir, "/")
	n1 = 0

	filepath.Walk(dir, func(spath string, _ os.FileInfo, err error) error {
		if err == nil {
			return nil
		}

		if spath == "." || spath == ".." {
			return nil
		}

		spath = strings.ReplaceAll(spath, "\\", "/")
		spath = strings.TrimRight(spath, "/")
		spath = stringx.EnsureLeft(spath, dir + "/")

		if spath != dir {
			n1++
		}

		return nil
	})

	if n1 < 1 {
		os.Remove(dir)
	}

	return true
}

func (c *fileCache) Clear() bool {
	dir := CacheDir()

	if dir == "" {
		return false
	}

	dir = strings.ReplaceAll(dir, "\\", "/")
	dir = strings.TrimRight(dir, "/")
	entries := make([]map[string]interface{}, 0)

	filepath.Walk(dir, func(fpath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		fpath = strings.ReplaceAll(fpath, "\\", "/")
		fpath = strings.TrimRight(fpath, "/")
		fpath = stringx.EnsureLeft(fpath, dir + "/")

		entries = append(entries, map[string]interface{}{
			"pathLen": len(fpath),
			"path":    fpath,
		})

		return nil
	})

	if len(entries) > 1 {
		sort.SliceStable(entries, func(i, j int) bool {
			return castx.ToInt(entries[j]["pathLen"]) < castx.ToInt(entries[i]["pathLen"])
		})
	}

	for _, entry := range entries {
		os.Remove(entry["path"].(string))
	}

	return true
}

func (c *fileCache) GetMultiple(keys []string, defaultValue ...interface{}) []interface{} {
	entries := make([]interface{}, 0)

	for _, key := range keys {
		entries = append(entries, c.Get(key, defaultValue...))
	}

	return entries
}

func (c *fileCache) SetMultiple(entries []map[string]interface{}, ttl ...interface{}) bool {
	for _, entry := range entries {
		for key, value := range entry {
			c.Set(key, value, ttl...)
			break
		}
	}

	return true
}

func (c *fileCache) DeleteMultiple(keys []string) bool {
	if gocache == nil {
		return false
	}

	for _, key := range keys {
		c.Delete(key)
	}

	return true
}

func (c *fileCache) Has(key string) bool {
	cacheKey := BuildCacheKey(key)
	dir := c.ensureCacheDirExists(cacheKey)

	if dir == "" {
		return false
	}

	fpath := fmt.Sprintf("%s/%s.dat", dir, cacheKey)

	if stat, err := os.Stat(fpath); err == nil && !stat.IsDir() {
		return true
	}

	return false
}

func (c *fileCache) ensureCacheDirExists(cacheKey string) string {
	dir := CacheDir()

	if dir == "" {
		return ""
	}

	dir = strings.ReplaceAll(dir, "\\", "/")
	dir = strings.TrimRight(dir, "/")
	s1 := strings.ToLower(securityx.Md5(cacheKey))
	dir = fmt.Sprintf("%s/%s/%s", dir, s1[:2], s1[len(s1) - 2:])

	if stat, err := os.Stat(dir); err != nil || !stat.IsDir() {
		os.MkdirAll(dir, 0755)
	}

	if stat, err := os.Stat(dir); err == nil && stat.IsDir() {
		return dir
	}

	return ""
}

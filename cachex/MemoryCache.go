package cachex

import (
	"github.com/meiguonet/mgboot-go-common/util/castx"
	"time"
)

type memoryCache struct {
}

func (c *memoryCache) Get(key string, defaultValue ...interface{}) interface{} {
	if gocache == nil {
		return nil
	}

	var dv interface{}

	if len(defaultValue) > 0 {
		dv = defaultValue[0]
	}

	cacheKey := BuildCacheKey(key)
	_entry, ok := gocache.Get(cacheKey)

	if !ok {
		return dv
	}

	entry, ok := _entry.(map[string]interface{})

	if !ok || len(entry) < 1 {
		return dv
	}

	data, ok := entry["data"]

	if !ok {
		return dv
	}

	if expireAt, ok := entry["expireAt"].(int64); ok && expireAt > 0 {
		if time.Now().Unix() > expireAt {
			c.Delete(key)
			return dv
		}
	}

	return data
}

func (c *memoryCache) Set(key string, value interface{}, ttl ...interface{}) bool {
	if gocache == nil {
		return false
	}

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

	cacheKey := BuildCacheKey(key)
	data := map[string]interface{}{"data": value}

	if _ttl > 0 {
		data["expireAt"] = time.Now().Add(_ttl).Unix()
		gocache.Set(cacheKey, data, _ttl)
	} else {
		gocache.SetDefault(cacheKey, data)
	}

	return true
}

func (c *memoryCache) Delete(key string) bool {
	if gocache == nil {
		return false
	}

	cacheKey := BuildCacheKey(key)
	gocache.Delete(cacheKey)
	return true
}

func (c *memoryCache) Clear() bool {
	if gocache == nil {
		return false
	}

	gocache.Flush()
	return true
}

func (c *memoryCache) GetMultiple(keys []string, defaultValue ...interface{}) []interface{} {
	if gocache == nil {
		return make([]interface{}, 0)
	}

	entries := make([]interface{}, 0)

	for _, key := range keys {
		entries = append(entries, c.Get(key, defaultValue...))
	}

	return entries
}

func (c *memoryCache) SetMultiple(entries []map[string]interface{}, ttl ...interface{}) bool {
	if gocache == nil {
		return false
	}

	for _, entry := range entries {
		for key, value := range entry {
			c.Set(key, value, ttl...)
			break
		}
	}

	return true
}

func (c *memoryCache) DeleteMultiple(keys []string) bool {
	if gocache == nil {
		return false
	}

	for _, key := range keys {
		c.Delete(key)
	}

	return true
}

func (c *memoryCache) Has(key string) bool {
	if gocache == nil {
		return false
	}

	cacheKey := BuildCacheKey(key)

	if _, ok := gocache.Get(cacheKey); ok {
		return true
	}

	return false
}

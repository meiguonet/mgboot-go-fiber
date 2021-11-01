package cachex

import (
	"context"
	"github.com/gomodule/redigo/redis"
	"github.com/meiguonet/mgboot-go-common/util/castx"
	"github.com/meiguonet/mgboot-go-common/util/jsonx"
	"github.com/meiguonet/mgboot-go-dal/poolx"
	"sync"
	"time"
)

type redisCache struct {
}

func (c *redisCache) Get(key string, defaultValue ...interface{}) interface{} {
	var dv interface{}

	if len(defaultValue) > 0 {
		dv = defaultValue[0]
	}

	conn, err := c.getRedisConn()

	if err != nil {
		return dv
	}

	defer conn.Close()
	cacheKey := BuildCacheKey(key)
	entry, err := redis.String(conn.Do("GET", cacheKey))

	if err != nil {
		return dv
	}

	map1 := jsonx.MapFrom(entry)

	if n1 := castx.ToInt64(map1["expireAt"]); n1 > 0 && time.Now().Unix() > n1 {
		c.Delete(key)
		return dv
	}

	if value, ok := map1["data"]; ok && value != nil {
		return value
	}

	return dv
}

func (c *redisCache) Set(key string, value interface{}, ttl ...interface{}) bool {
	var err error
	conn, err := c.getRedisConn()

	if err != nil {
		return false
	}

	defer conn.Close()
	cacheKey := BuildCacheKey(key)
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

	entry := map[string]interface{}{"data": value}

	if _ttl > 0 {
		entry["expireAt"] = time.Now().Add(_ttl).Unix()
		_, err = conn.Do("SET", cacheKey, jsonx.ToJson(entry))

		if err == nil {
			_, _ = conn.Do("PEXPIRE", cacheKey, _ttl.Milliseconds())
		}
	} else {
		_, err = conn.Do("SET", cacheKey, jsonx.ToJson(entry))
	}

	return err == nil
}

func (c *redisCache) Delete(key string) bool {
	conn, err := c.getRedisConn()

	if err != nil {
		return false
	}

	defer conn.Close()
	cacheKey := BuildCacheKey(key)
	_, err = conn.Do("DEL", cacheKey)
	return err == nil
}

func (c *redisCache) Clear() bool {
	return true
}

func (c *redisCache) GetMultiple(keys []string, defaultValue ...interface{}) []interface{} {
	if len(keys) < 1 {
		return make([]interface{}, 0)
	}

	var dv interface{}

	if len(defaultValue) > 0 {
		dv = defaultValue[0]
	}

	entries :=  make([]interface{}, 0, len(keys))

	for i := 0; i < len(keys); i++ {
		entries = append(entries, dv)
	}

	conn, err := c.getRedisConn()

	if err != nil {
		return entries
	}

	defer conn.Close()
	wg := sync.WaitGroup{}
	wg.Add(len(keys))
	mu := sync.RWMutex{}

	for idx, key := range keys {
		go func(idx int, key string) {
			defer wg.Done()
			cacheKey := BuildCacheKey(key)
			entry, err := redis.String(conn.Do("GET", cacheKey))

			if err != nil {
				mu.Lock()
				entries[idx] = dv
				mu.Unlock()
				return
			}

			map1 := jsonx.MapFrom(entry)

			if n1 := castx.ToInt64(map1["expireAt"]); n1 > 0 && time.Now().Unix() > n1 {
				_, _ = conn.Do("DEL", cacheKey)
				mu.Lock()
				entries[idx] = dv
				mu.Unlock()
				return
			}

			if value, ok := map1["data"]; ok && value != nil {
				mu.Lock()
				entries[idx] = value
				mu.Unlock()
				return
			}

			mu.Lock()
			entries[idx] = dv
			mu.Unlock()
		}(idx, key)
	}

	wg.Wait()
	return entries
}

func (c *redisCache) SetMultiple(entries []map[string]interface{}, ttl ...interface{}) bool {
	conn, err := c.getRedisConn()

	if err != nil {
		return false
	}

	defer conn.Close()
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

	wg := sync.WaitGroup{}
	wg.Add(len(entries))

	for _, entry := range entries {
		go func(entry map[string]interface{}) {
			defer wg.Done()
			var key string
			var value interface{}

			for k, v := range entry {
				key = k
				value = v
				break
			}

			cacheKey := BuildCacheKey(key)
			payload := map[string]interface{}{"data": value}

			if _ttl > 0 {
				payload["expireAt"] = time.Now().Add(_ttl).Unix()
				_, err = conn.Do("SET", cacheKey, jsonx.ToJson(payload))

				if err == nil {
					_, _ = conn.Do("PEXPIRE", cacheKey, _ttl.Milliseconds())
				}
			} else {
				_, _ = conn.Do("SET", cacheKey, jsonx.ToJson(payload))
			}
		}(entry)
	}

	wg.Wait()
	return true
}

func (c *redisCache) DeleteMultiple(keys []string) bool {
	if len(keys) < 1 {
		return true
	}

	conn, err := c.getRedisConn()

	if err != nil {
		return false
	}

	defer conn.Close()
	wg := sync.WaitGroup{}
	wg.Add(len(keys))

	for _, key := range keys {
		go func(key string) {
			defer wg.Done()
			cacheKey := BuildCacheKey(key)
			_, _ = conn.Do("DEL", cacheKey)
		}(key)
	}

	wg.Wait()
	return true
}

func (c *redisCache) Has(key string) bool {
	conn, err := c.getRedisConn()

	if err != nil {
		return false
	}

	defer conn.Close()
	cacheKey := BuildCacheKey(key)
	exists, _ := redis.Bool(conn.Do("EXISTS", cacheKey))
	return exists
}

func (c *redisCache) getRedisConn() (redis.Conn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2 * time.Second)
	defer cancel()
	return poolx.GetRedisConnection(ctx)
}

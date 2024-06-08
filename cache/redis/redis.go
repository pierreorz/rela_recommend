// Copyright 2014 beego Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package redis for cache provider
package redis

import (
	"rela_recommend/cache"
	"time"

	"github.com/garyburd/redigo/redis"
)

// Cache is Redis cache adapter.
type Cache struct {
	conn     *redis.Pool // redis connection pool
	conninfo string
	dbNum    int
	password string
}

// NewRedisCache create new redis cache with default collection name.
func NewRedisCache(conninfo, password string, dbNum int) (cache.Cache, error) {
	var cache Cache
	cache.dbNum = dbNum
	cache.conninfo = conninfo
	cache.password = password

	cache.connectInit()

	return &cache, nil
}

// actually do the redis cmds
func (rc *Cache) do(commandName string, args ...interface{}) (reply interface{}, err error) {
	c := rc.conn.Get()
	defer c.Close()

	return c.Do(commandName, args...)
}

// Get cache from redis.
func (rc *Cache) Get(key string) (interface{}, error) {
	return rc.do("GET", key)
}

func (rc *Cache) Mget(keys []string) ([]interface{}, error) {
	if len(keys) == 0 {
		return make([]interface{}, 0), nil
	}
	args := make([]interface{}, 0)
	for _, key := range(keys) {
		args = append(args, key)
	}
	rv, err := redis.Values(rc.do("MGET", args...))
	return rv, err
}

func (rc *Cache) LRange(key string, start int, end int) ([][]byte, error) {
	rv, err := redis.ByteSlices(rc.do("LRANGE", key, start, end))
	return rv, err
}

func (rc *Cache)ZRange(key string,start int ,end int)([]interface{},error){
	rv,err :=redis.Values(rc.do("ZRANGE",key,start,end))
	return rv,err
}

func (rc *Cache) SMembers(key string) ([]interface{}, error) {
	rv, err := redis.Values(rc.do("SMEMBERS", key))
	return rv, err
}

// GetMulti get cache from redis.
func (rc *Cache) MultiGet(keys []string) []interface{} {
	size := len(keys)
	var rv []interface{}
	c := rc.conn.Get()
	defer c.Close()
	var err error
	for _, key := range keys {
		err = c.Send("GET", key)
		if err != nil {
			goto ERROR
		}
	}
	if err = c.Flush(); err != nil {
		goto ERROR
	}
	for i := 0; i < size; i++ {
		if v, err := c.Receive(); err == nil {
			rv = append(rv, v.([]byte))
		} else {
			rv = append(rv, err)
		}
	}
	return rv
ERROR:
	rv = rv[0:0]
	for i := 0; i < size; i++ {
		rv = append(rv, nil)
	}

	return rv
}

func (rc *Cache) MsetEx(keyValMap map[string]interface{}, expire int64) error {
	if len(keyValMap) == 0 {
		return nil
	}
	for key, val := range keyValMap {
		_, err := rc.do("SETEX", key, expire, val)
		if err != nil {
			return err
		}
	}
	return nil
}

// Put put cache to redis.
func (rc *Cache) SetEx(key string, val interface{}, expireSeconds int) error {
	_, err := rc.do("SETEX", key, expireSeconds, val)
	return err
}

// Set put cache to redis.
func (rc *Cache) Set(key string, val interface{}) error {
	_, err := rc.do("SET", key, val)
	return err
}

// Delete delete cache in redis.
func (rc *Cache) Del(key string) error {
	_, err := rc.do("DEL", key)
	return err
}

// IsExist check cache's existence in redis.
func (rc *Cache) Exists(key string) bool {
	v, err := redis.Bool(rc.do("EXISTS", key))
	if err != nil {
		return false
	}
	return v
}

// Incr increase counter in redis.
func (rc *Cache) Incr(key string) (int64, error) {
	return redis.Int64(rc.do("INCR", key))
}

// IncrBy increase counter in redis.
func (rc *Cache) IncrBy(key string, count int) (int64, error) {
	return redis.Int64(rc.do("INCRBY", key, count))
}

// Decr decrease counter in redis.
func (rc *Cache) Decr(key string) (int64, error) {
	return redis.Int64(rc.do("DECR", key))
}

// get redis pool
func (rc *Cache) GetConn() interface{} {
	return rc.conn
}

func (rc *Cache) Close() {
	rc.conn.Close()
}

// connect to redis.
func (rc *Cache) connectInit() {
	dialFunc := func() (c redis.Conn, err error) {
		c, err = redis.Dial("tcp", rc.conninfo,
			redis.DialConnectTimeout(1000 * time.Millisecond),
			redis.DialReadTimeout(500 * time.Millisecond),
			redis.DialWriteTimeout(1000 * time.Millisecond))
		if err != nil {
			return nil, err
		}

		if rc.password != "" {
			if _, err := c.Do("AUTH", rc.password); err != nil {
				c.Close()
				return nil, err
			}
		}

		_, selecterr := c.Do("SELECT", rc.dbNum)
		if selecterr != nil {
			c.Close()
			return nil, selecterr
		}
		return
	}
	// initialize a new pool
	rc.conn = &redis.Pool{
		MaxIdle:     20,
		MaxActive:   1250,
		Wait:        true,
		IdleTimeout: 180 * time.Second,
		Dial:        dialFunc,
	}
}

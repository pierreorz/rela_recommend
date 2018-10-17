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
	"theL_api_golang/cache"
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

func (rc *Cache) Mget(keys []string) []interface{} {
	if len(keys) == 0 {
		return nil
	}
	size := len(keys)
	var rv []interface{}
	c := rc.conn.Get()
	defer c.Close()
	for _, key := range keys {
		c.Send("GET", key)
	}
	c.Flush()
	for i := 0; i < size; i++ {
		if v, err := c.Receive(); err == nil {
			if v != nil {
				rv = append(rv, v.([]byte))
			}
		} else {
			rv = append(rv, err)
		}
	}
	return rv
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

// connect to redis.
func (rc *Cache) connectInit() {
	dialFunc := func() (c redis.Conn, err error) {
		c, err = redis.Dial("tcp", rc.conninfo)
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

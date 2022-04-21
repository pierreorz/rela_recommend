package redisCluster

import (
	redisCluster "github.com/chasex/redis-go-cluster"
	"github.com/garyburd/redigo/redis"
	"rela_recommend/cache"
	"rela_recommend/log"
	"strings"
	"time"
)

// Cache is Redis cache adapter.
type Cache struct {
	conn     *redisCluster.Cluster // redis connection pool
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
	c := rc.conn
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
	var rv []interface{}
	c := rc.conn
	batch := c.NewBatch()
	for _, key := range keys {
		batch.Put("GET", key)
	}
	rv, err := c.RunBatch(batch)
	return rv, err
}

func (rc *Cache) LRange(key string, start int, end int) ([][]byte, error) {
	rv, err := redis.ByteSlices(rc.do("LRANGE", key, start, end))
	return rv, err
}

func (rc *Cache) ZRange(key string, start int, end int) ([]interface{}, error) {
	rv, err := redis.Values(rc.do("ZRANGE", key, start, end))
	return rv, err
}
func (rc *Cache) SMembers(key string) ([]interface{}, error) {
	rv, err := redis.Values(rc.do("SMEMBERS", key))
	return rv, err
}

func (rc *Cache) MsetEx(keyValMap map[string]interface{}, expire int64) error {
	if len(keyValMap) == 0 {
		return nil
	}
	c := rc.conn
	batch := c.NewBatch()
	for key, val := range keyValMap {
		batch.Put("SETEX", key, expire, val)
	}
	_, err := c.RunBatch(batch)
	return err
}


// GetMulti get cache from redis.
func (rc *Cache) MultiGet(keys []string) []interface{} {
	size := len(keys)
	var rv []interface{}
	c := rc.conn
	var err error
	batch := c.NewBatch()
	for _, key := range keys {
		err = batch.Put("GET", key)
		if err != nil {
			goto ERROR
		}
	}
	rv, err = c.RunBatch(batch)
	if err != nil {
		log.Error(err)
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

func (rc *Cache) Close() {
	rc.conn.Close()
}

// connect to redis.
func (rc *Cache) connectInit() {
	var err error
	addr := strings.Split(rc.conninfo, ",")
	rc.conn, err = redisCluster.NewCluster(
		&redisCluster.Options{
			StartNodes:   addr,
			ConnTimeout:  1000 * time.Millisecond,
			ReadTimeout:  500 * time.Millisecond,
			WriteTimeout: 1000 * time.Millisecond,
			KeepAlive:    1000,
			AliveTime:    180 * time.Second,
		})
	if rc.conn == nil {
		log.Error("redis cluster init error!")
		return
	}
	if rc.password != "" {
		if _, err := rc.conn.Do("AUTH", rc.password); err != nil {
			rc.conn.Close()
			log.Error(err)
		}
	}
	if err != nil {
		log.Error(err)
	}

	_, selecterr := rc.conn.Do("SELECT", rc.dbNum)
	if selecterr != nil {
		rc.conn.Close()
		log.Error(selecterr)
	}
}

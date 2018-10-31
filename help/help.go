package help

import (
	"encoding/json"
	"errors"
	"github.com/garyburd/redigo/redis"
	"github.com/ssdb/gossdb/ssdb"
	"rela_recommend/cache"
	"rela_recommend/utils"
	"time"
)

var errNoData = errors.New("errNoData")

func ExpireAt(cach cache.Cache, key string, t time.Time) error {
	conn := cach.GetConn()
	switch result := conn.(type) {
	case *redis.Pool:
		return expireAtByRedis(result, key, t)
	case *ssdb.Client:
		return nil
	default:
		return nil
	}
}

func expireAtByRedis(conn *redis.Pool, key string, t time.Time) error {
	c := conn.Get()
	defer c.Close()
	_, err := c.Do("EXPIREAT", key, t.Unix())
	return err
}

func Expire(cach cache.Cache, key string, expireSeconds int) error {
	conn := cach.GetConn()
	switch result := conn.(type) {
	case *redis.Pool:
		return expireByRedis(result, key, expireSeconds)
	case *ssdb.Client:
		return nil
	default:
		return nil
	}
}

func expireByRedis(conn *redis.Pool, key string, expireSeconds int) error {
	c := conn.Get()
	defer c.Close()

	_, err := c.Do("EXPIRE", key, expireSeconds)
	return err
}

func GetStructByCache(cach cache.Cache, key string, i interface{}) error {
	value, _ := cach.Get(key)
	if value == nil {
		return errNoData
	}
	data := utils.GetBytes(value)
	return json.Unmarshal(data, i)
}

func SetExStructByCache(cach cache.Cache, key string, value interface{}, expireSeconds int) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return cach.SetEx(key, data, expireSeconds)
}

func SetStructByCache(cach cache.Cache, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return cach.Set(key, data)
}

func SetBytesByCache(cach cache.Cache, key string, value []byte) error {
	return cach.Set(key, value)
}

func IncrBy(cach cache.Cache, key string, increment int64) (int64, error) {
	conn := cach.GetConn()
	switch result := conn.(type) {
	case *redis.Pool:
		return incrBy(result, key, increment)
	case *ssdb.Client:
		return 0, nil
	default:
		return 0, nil
	}
}

func incrBy(conn *redis.Pool, key string, increment int64) (int64, error) {
	c := conn.Get()
	defer c.Close()
	return redis.Int64(c.Do("INCRBY", key, increment))
}

func SisMember(cach cache.Cache, key string, member interface{}) (bool, error) {
	conn := cach.GetConn()
	switch result := conn.(type) {
	case *redis.Pool:
		return sisMemberByRedis(result, key, member)
	case *ssdb.Client:
		return false, nil
	default:
		return false, nil
	}
}

func sisMemberByRedis(conn *redis.Pool, key string, member interface{}) (bool, error) {
	c := conn.Get()
	defer c.Close()
	return redis.Bool(c.Do("SISMEMBER", key, member))
}

func Sadd(cach cache.Cache, key string, member interface{}) error {
	conn := cach.GetConn()
	switch result := conn.(type) {
	case *redis.Pool:
		return sAddByRedis(result, key, member)
	case *ssdb.Client:
		return nil
	default:
		return nil
	}
}

func sAddByRedis(conn *redis.Pool, key string, member interface{}) error {
	c := conn.Get()
	defer c.Close()
	_, err := c.Do("SADD", key, member)
	return err
}

func Srem(cach cache.Cache, key string, member interface{}) error {
	conn := cach.GetConn()
	switch result := conn.(type) {
	case *redis.Pool:
		return sRemByRedis(result, key, member)
	case *ssdb.Client:
		return nil
	default:
		return nil
	}
}

func sRemByRedis(conn *redis.Pool, key string, member interface{}) error {
	c := conn.Get()
	defer c.Close()
	_, err := c.Do("SREM", key, member)
	return err
}

func SmembersInt64s(cach cache.Cache, key string) ([]int64, error) {
	conn := cach.GetConn()
	switch result := conn.(type) {
	case *redis.Pool:
		return sMembersInt64sByRedis(result, key)
	case *ssdb.Client:
		return nil, nil
	default:
		return nil, nil
	}
}

func sMembersInt64sByRedis(conn *redis.Pool, key string) (int64s []int64, err error) {
	c := conn.Get()
	defer c.Close()
	var values []interface{}
	values, err = redis.Values(c.Do("SMEMBERS", key))
	if err != nil {
		return nil, err
	}
	if err = redis.ScanSlice(values, &int64s); err != nil {
		return nil, err
	}
	return int64s, nil
}

func SmembersStrings(cach cache.Cache, key string) ([]string, error) {
	conn := cach.GetConn()
	switch result := conn.(type) {
	case *redis.Pool:
		return sMembersStringsByRedis(result, key)
	case *ssdb.Client:
		return nil, nil
	default:
		return nil, nil
	}
}

func sMembersStringsByRedis(conn *redis.Pool, key string) (vals []string, err error) {
	c := conn.Get()
	defer c.Close()
	var values []interface{}
	values, err = redis.Values(c.Do("SMEMBERS", key))
	if err != nil {
		return nil, err
	}
	if err = redis.ScanSlice(values, &vals); err != nil {
		return nil, err
	}
	return vals, nil
}

func Scard(cach cache.Cache, key string) (int, error) {
	conn := cach.GetConn()
	switch result := conn.(type) {
	case *redis.Pool:
		return sCardByRedis(result, key)
	case *ssdb.Client:
		return 0, nil
	default:
		return 0, nil
	}
}

func sCardByRedis(conn *redis.Pool, key string) (int, error) {
	c := conn.Get()
	defer c.Close()
	return redis.Int(c.Do("SCARD", key))
}

func SrandMemberStrings(cach cache.Cache, key string, count int) ([]string, error) {
	conn := cach.GetConn()
	switch result := conn.(type) {
	case *redis.Pool:
		return sRandMemberStringsByRedis(result, key, count)
	case *ssdb.Client:
		return nil, nil
	default:
		return nil, nil
	}
}

func sRandMemberStringsByRedis(conn *redis.Pool, key string, count int) ([]string, error) {
	c := conn.Get()
	defer c.Close()
	return redis.Strings(c.Do("SRANDMEMBER", key, count))
}

func SrandMemberInt64s(cach cache.Cache, key string, count int) ([]int64, error) {
	conn := cach.GetConn()
	switch result := conn.(type) {
	case *redis.Pool:
		return sRandMemberInt64sByRedis(result, key, count)
	case *ssdb.Client:
		return nil, nil
	default:
		return nil, nil
	}
}

func sRandMemberInt64sByRedis(conn *redis.Pool, key string, count int) (int64s []int64, err error) {
	c := conn.Get()
	defer c.Close()
	var values []interface{}
	values, err = redis.Values(c.Do("SRANDMEMBER", key, count))
	if err != nil {
		return nil, err
	}
	if err = redis.ScanSlice(values, &int64s); err != nil {
		return nil, err
	}
	return int64s, nil
}

func SinterInt64s(cach cache.Cache, keys ...interface{}) ([]int64, error) {
	conn := cach.GetConn()
	switch result := conn.(type) {
	case *redis.Pool:
		return sinterInt64sByRedis(result, keys...)
	case *ssdb.Client:
		return nil, nil
	default:
		return nil, nil
	}
}

func sinterInt64sByRedis(conn *redis.Pool, keys ...interface{}) (int64s []int64, err error) {
	c := conn.Get()
	defer c.Close()
	var values []interface{}
	values, err = redis.Values(c.Do("SINTER", keys...))
	if err != nil {
		return nil, err
	}
	if err = redis.ScanSlice(values, &int64s); err != nil {
		return nil, err
	}
	return int64s, nil
}

func ZrangeStrings(cach cache.Cache, key string, start interface{}, end interface{}) ([]string, error) {
	conn := cach.GetConn()
	switch result := conn.(type) {
	case *redis.Pool:
		return zRangeStringsByRedis(result, key, start, end)
	case *ssdb.Client:
		return nil, nil
	default:
		return nil, nil
	}
}

func zRangeStringsByRedis(conn *redis.Pool, key string, start interface{}, end interface{}) (strings []string, err error) {
	c := conn.Get()
	defer c.Close()
	return redis.Strings(c.Do("ZRANGE", key, start, end))
}

func ZrevrangeStrings(cach cache.Cache, key string, start interface{}, end interface{}) ([]string, error) {
	conn := cach.GetConn()
	switch result := conn.(type) {
	case *redis.Pool:
		return zRevrangeStringsByRedis(result, key, start, end)
	case *ssdb.Client:
		return nil, nil
	default:
		return nil, nil
	}
}

func zRevrangeStringsByRedis(conn *redis.Pool, key string, start interface{}, end interface{}) (strings []string, err error) {
	c := conn.Get()
	defer c.Close()
	return redis.Strings(c.Do("ZREVRANGE", key, start, end))
}

func ZrevrangeScoreInt64Map(cach cache.Cache, key string, start interface{}, end interface{}) (map[string]int64, error) {
	conn := cach.GetConn()
	switch result := conn.(type) {
	case *redis.Pool:
		return zRevrangeScoreInt64MapByRedis(result, key, start, end)
	case *ssdb.Client:
		return nil, nil
	default:
		return nil, nil
	}
}

func zRevrangeScoreInt64MapByRedis(conn *redis.Pool, key string, start interface{}, end interface{}) (map[string]int64, error) {
	c := conn.Get()
	defer c.Close()
	return redis.Int64Map(c.Do("ZREVRANGE", key, start, end, "WITHSCORES"))
}

func ZrevrangeScoreStrings(cach cache.Cache, key string, start interface{}, end interface{}) ([]string, error) {
	conn := cach.GetConn()
	switch result := conn.(type) {
	case *redis.Pool:
		return zRevrangeScoreStringsByRedis(result, key, start, end)
	case *ssdb.Client:
		return nil, nil
	default:
		return nil, nil
	}
}

func zRevrangeScoreStringsByRedis(conn *redis.Pool, key string, start interface{}, end interface{}) ([]string, error) {
	c := conn.Get()
	defer c.Close()
	return redis.Strings(c.Do("ZREVRANGE", key, start, end, "WITHSCORES"))
}

func ZrangeInt64s(cach cache.Cache, key string, start interface{}, end interface{}) (int64s []int64, err error) {
	conn := cach.GetConn()
	switch result := conn.(type) {
	case *redis.Pool:
		return zrangeInt64sByRedis(result, key, start, end)
	case *ssdb.Client:
		return nil, nil
	default:
		return nil, nil
	}
}

// true values, err = redis.Values(c.Do("ZREVRANGE", key, start, end))
func zrangeInt64sByRedis(conn *redis.Pool, key string, start interface{}, end interface{}) (int64s []int64, err error) {
	c := conn.Get()
	defer c.Close()
	var values []interface{}
	values, err = redis.Values(c.Do("ZRANGE", key, start, end))
	if err != nil {
		return nil, err
	}
	if err = redis.ScanSlice(values, &int64s); err != nil {
		return nil, err
	}
	return int64s, nil
}

func ZrangeInt64sWithScores(cach cache.Cache, key string, start interface{}, end interface{}) (int64s []int64, err error) {
	conn := cach.GetConn()
	switch result := conn.(type) {
	case *redis.Pool:
		return zrangeInt64sWithScoresByRedis(result, key, start, end)
	case *ssdb.Client:
		return nil, nil
	default:
		return nil, nil
	}
}

// true values, err = redis.Values(c.Do("ZREVRANGE", key, start, end))
func zrangeInt64sWithScoresByRedis(conn *redis.Pool, key string, start interface{}, end interface{}) (int64s []int64, err error) {
	c := conn.Get()
	defer c.Close()
	var values []interface{}
	values, err = redis.Values(c.Do("ZRANGE", key, start, end))
	if err != nil {
		return nil, err
	}
	if err = redis.ScanSlice(values, &int64s); err != nil {
		return nil, err
	}
	return int64s, nil
}

func ZrevRangeInt64s(cach cache.Cache, key string, start interface{}, end interface{}) (int64s []int64, err error) {
	conn := cach.GetConn()
	switch result := conn.(type) {
	case *redis.Pool:
		return zRevRangeInt64sByRedis(result, key, start, end)
	case *ssdb.Client:
		return nil, nil
	default:
		return nil, nil
	}
}

func zRevRangeInt64sByRedis(conn *redis.Pool, key string, start interface{}, end interface{}) (int64s []int64, err error) {
	c := conn.Get()
	defer c.Close()
	var values []interface{}
	values, err = redis.Values(c.Do("ZREVRANGE", key, start, end))
	if err != nil {
		return nil, err
	}
	if err = redis.ScanSlice(values, &int64s); err != nil {
		return nil, err
	}
	return int64s, nil
}

func ZcardInt(cach cache.Cache, key string) (int, error) {
	conn := cach.GetConn()
	switch result := conn.(type) {
	case *redis.Pool:
		return zCardIntCountByRedis(result, key)
	case *ssdb.Client:
		return 0, nil
	default:
		return 0, nil
	}
}

func zCardIntCountByRedis(conn *redis.Pool, key string) (int, error) {
	c := conn.Get()
	defer c.Close()
	return redis.Int(c.Do("ZCARD", key))
}

func ZscoreInt(cach cache.Cache, key string, member interface{}) (int, error) {
	conn := cach.GetConn()
	switch result := conn.(type) {
	case *redis.Pool:
		return zscoreIntByRedis(result, key, member)
	case *ssdb.Client:
		return 0, nil
	default:
		return 0, nil
	}
}

func zscoreIntByRedis(conn *redis.Pool, key string, member interface{}) (int, error) {
	c := conn.Get()
	defer c.Close()
	return redis.Int(c.Do("ZSCORE", key, member))
}

func ZscoreInt64(cach cache.Cache, key string, member interface{}) (int64, error) {
	conn := cach.GetConn()
	switch result := conn.(type) {
	case *redis.Pool:
		return zscoreInt64ByRedis(result, key, member)
	case *ssdb.Client:
		return 0, nil
	default:
		return 0, nil
	}
}

func zscoreInt64ByRedis(conn *redis.Pool, key string, member interface{}) (int64, error) {
	c := conn.Get()
	defer c.Close()
	return redis.Int64(c.Do("ZSCORE", key, member))
}

func Zadd(cach cache.Cache, key string, score interface{}, member interface{}) (int, error) {
	conn := cach.GetConn()
	switch result := conn.(type) {
	case *redis.Pool:
		return zAddByRedis(result, key, score, member)
	case *ssdb.Client:
		return 0, nil
	default:
		return 0, nil
	}
}

func zAddByRedis(conn *redis.Pool, key string, score interface{}, member interface{}) (int, error) {
	c := conn.Get()
	defer c.Close()
	return redis.Int(c.Do("ZADD", key, score, member))
}

func ZrangeByScoreInt64s(cach cache.Cache, key string, min interface{}, max interface{}) (int64s []int64, err error) {
	conn := cach.GetConn()
	switch result := conn.(type) {
	case *redis.Pool:
		values, err := redis.Values(zRangeByScoreByRedis(result, key, min, max))
		if err != nil {
			return nil, err
		}
		if err = redis.ScanSlice(values, &int64s); err != nil {
			return nil, err
		}
		return int64s, nil
	case *ssdb.Client:
		return nil, nil
	default:
		return nil, nil
	}
}

func ZrangeByScoreStrings(cach cache.Cache, key string, min interface{}, max interface{}) ([]string, error) {
	conn := cach.GetConn()
	switch result := conn.(type) {
	case *redis.Pool:
		return redis.Strings(zRangeByScoreByRedis(result, key, min, max))
	case *ssdb.Client:
		return nil, nil
	default:
		return nil, nil
	}
}

func ZrangeByScore(cach cache.Cache, key string, min interface{}, max interface{}) (interface{}, error) {
	conn := cach.GetConn()
	switch result := conn.(type) {
	case *redis.Pool:
		return zRangeByScoreByRedis(result, key, min, max)
	case *ssdb.Client:
		return nil, nil
	default:
		return nil, nil
	}
}

func zRangeByScoreByRedis(conn *redis.Pool, key string, min interface{}, max interface{}) (interface{}, error) {
	c := conn.Get()
	defer c.Close()
	return c.Do("ZRANGEBYSCORE", key, min, max)
}

func Zrem(cach cache.Cache, key string, member interface{}) error {
	conn := cach.GetConn()
	switch result := conn.(type) {
	case *redis.Pool:
		return zRemByRedis(result, key, member)
	case *ssdb.Client:
		return nil
	default:
		return nil
	}
}

func zRemByRedis(conn *redis.Pool, key string, member interface{}) error {
	c := conn.Get()
	defer c.Close()
	_, err := c.Do("ZREM", key, member)
	return err
}

func ZremRangeByRank(cach cache.Cache, key string, start, stop int) error {
	conn := cach.GetConn()
	switch result := conn.(type) {
	case *redis.Pool:
		return zRemRangeByRankByRedis(result, key, start, stop)
	case *ssdb.Client:
		return nil
	default:
		return nil
	}
}

func zRemRangeByRankByRedis(conn *redis.Pool, key string, start, stop int) error {
	c := conn.Get()
	defer c.Close()
	_, err := c.Do("ZREMRANGEBYRANK", key, start, stop)
	return err
}

func Zcount(cach cache.Cache, key string, start, stop interface{}) (int, error) {
	conn := cach.GetConn()
	switch result := conn.(type) {
	case *redis.Pool:
		return zCountByRedis(result, key, start, stop)
	case *ssdb.Client:
		return 0, nil
	default:
		return 0, nil
	}
}

func zCountByRedis(conn *redis.Pool, key string, start, stop interface{}) (int, error) {
	c := conn.Get()
	defer c.Close()
	return redis.Int(c.Do("ZCOUNT", key, start, stop))
}

func LrangeInt64s(cach cache.Cache, key string, start int, stop int) ([]int64, error) {
	conn := cach.GetConn()
	switch result := conn.(type) {
	case *redis.Pool:
		return lrangeInt64sByRedis(result, key, start, stop)
	case *ssdb.Client:
		return nil, nil
	default:
		return nil, nil
	}
}

func lrangeInt64sByRedis(conn *redis.Pool, key string, start int, stop int) (int64s []int64, err error) {
	c := conn.Get()
	defer c.Close()
	var values []interface{}
	values, err = redis.Values(c.Do("LRANGE", key, start, stop))
	if err != nil {
		return nil, err
	}
	if err = redis.ScanSlice(values, &int64s); err != nil {
		return nil, err
	}
	return int64s, nil
}

func LrangeStrings(cach cache.Cache, key string, start int, stop int) ([]string, error) {
	conn := cach.GetConn()
	switch result := conn.(type) {
	case *redis.Pool:
		return lrangeStringsByRedis(result, key, start, stop)
	case *ssdb.Client:
		return nil, nil
	default:
		return nil, nil
	}
}

func lrangeStringsByRedis(conn *redis.Pool, key string, start int, stop int) (strings []string, err error) {
	c := conn.Get()
	defer c.Close()
	return redis.Strings(c.Do("LRANGE", key, start, stop))
}

func Llen(cach cache.Cache, key string) (int, error) {
	conn := cach.GetConn()
	switch result := conn.(type) {
	case *redis.Pool:
		return llenByRedis(result, key)
	case *ssdb.Client:
		return 0, nil
	default:
		return 0, nil
	}
}

func llenByRedis(conn *redis.Pool, key string) (int, error) {
	c := conn.Get()
	defer c.Close()
	return redis.Int(c.Do("LLEN", key))
}

func Rpush(cach cache.Cache, key string, value interface{}) error {
	conn := cach.GetConn()
	switch result := conn.(type) {
	case *redis.Pool:
		return rPushbyRedis(result, key, value)
	case *ssdb.Client:
		return nil
	default:
		return nil
	}
}

func rPushbyRedis(conn *redis.Pool, key string, value interface{}) error {
	c := conn.Get()
	defer c.Close()
	_, err := c.Do("RPUSH", key, value)
	return err
}

func Lpush(cach cache.Cache, key string, value interface{}) error {
	conn := cach.GetConn()
	switch result := conn.(type) {
	case *redis.Pool:
		return lPushbyRedis(result, key, value)
	case *ssdb.Client:
		return nil
	default:
		return nil
	}
}

func lPushbyRedis(conn *redis.Pool, key string, value interface{}) error {
	c := conn.Get()
	defer c.Close()
	_, err := c.Do("LPUSH", key, value)
	return err
}

func LpushX(cach cache.Cache, key string, value interface{}) error {
	conn := cach.GetConn()
	switch result := conn.(type) {
	case *redis.Pool:
		return lPushXbyRedis(result, key, value)
	case *ssdb.Client:
		return nil
	default:
		return nil
	}
}

func lPushXbyRedis(conn *redis.Pool, key string, value interface{}) error {
	c := conn.Get()
	defer c.Close()
	_, err := c.Do("LPUSHX", key, value)
	return err
}

func RpushX(cach cache.Cache, key string, value interface{}) error {
	conn := cach.GetConn()
	switch result := conn.(type) {
	case *redis.Pool:
		return rPushXbyRedis(result, key, value)
	case *ssdb.Client:
		return nil
	default:
		return nil
	}
}

func rPushXbyRedis(conn *redis.Pool, key string, value interface{}) error {
	c := conn.Get()
	defer c.Close()
	_, err := c.Do("RPUSHX", key, value)
	return err
}

func Rpop(cach cache.Cache, key string) (interface{}, error) {
	conn := cach.GetConn()
	switch result := conn.(type) {
	case *redis.Pool:
		return rPopbyRedis(result, key)
	case *ssdb.Client:
		return nil, nil
	default:
		return nil, nil
	}
}

func rPopbyRedis(conn *redis.Pool, key string) (interface{}, error) {
	c := conn.Get()
	defer c.Close()
	return c.Do("RPOP", key)
}

func Lpop(cach cache.Cache, key string) (interface{}, error) {
	conn := cach.GetConn()
	switch result := conn.(type) {
	case *redis.Pool:
		return lPopbyRedis(result, key)
	case *ssdb.Client:
		return nil, nil
	default:
		return nil, nil
	}
}

func lPopbyRedis(conn *redis.Pool, key string) (interface{}, error) {
	c := conn.Get()
	defer c.Close()
	return c.Do("LPOP", key)
}

func Lrem(cach cache.Cache, key string, count int, value interface{}) error {
	conn := cach.GetConn()
	switch result := conn.(type) {
	case *redis.Pool:
		return lRemByRedis(result, key, count, value)
	case *ssdb.Client:
		return nil
	default:
		return nil
	}
}

func lRemByRedis(conn *redis.Pool, key string, count int, value interface{}) error {
	c := conn.Get()
	defer c.Close()
	_, err := c.Do("LREM", key, count, value)
	return err
}

func Ltrim(cach cache.Cache, key string, start, stop int) error {
	conn := cach.GetConn()
	switch result := conn.(type) {
	case *redis.Pool:
		return lTrimByRedis(result, key, start, stop)
	case *ssdb.Client:
		return nil
	default:
		return nil
	}
}

func lTrimByRedis(conn *redis.Pool, key string, start, stop int) error {
	c := conn.Get()
	defer c.Close()
	_, err := c.Do("LTRIM", key, start, stop)
	return err
}

func Hset(cach cache.Cache, key string, field string, value interface{}) error {
	conn := cach.GetConn()
	switch result := conn.(type) {
	case *redis.Pool:
		return hSetByRedis(result, key, field, value)
	case *ssdb.Client:
		return nil
	default:
		return nil
	}
}

func hSetByRedis(conn *redis.Pool, key string, field string, value interface{}) error {
	c := conn.Get()
	defer c.Close()
	_, err := c.Do("HSET", key, field, value)
	return err
}

func Hdel(cach cache.Cache, key string, field interface{}) error {
	conn := cach.GetConn()
	switch result := conn.(type) {
	case *redis.Pool:
		return hDelByRedis(result, key, field)
	case *ssdb.Client:
		return nil
	default:
		return nil
	}
}

func hDelByRedis(conn *redis.Pool, key string, field interface{}) error {
	c := conn.Get()
	defer c.Close()
	_, err := c.Do("HDEL", key, field)
	return err
}

func HgetInt64(cach cache.Cache, key string, field string) (int64, error) {
	conn := cach.GetConn()
	switch result := conn.(type) {
	case *redis.Pool:
		return hGetInt64ByRedis(result, key, field)
	case *ssdb.Client:
		return 0, nil
	default:
		return 0, nil
	}
}

func hGetInt64ByRedis(conn *redis.Pool, key string, field string) (int64, error) {
	c := conn.Get()
	defer c.Close()
	return redis.Int64(c.Do("HGET", key, field))
}

func HgetInt(cach cache.Cache, key string, field interface{}) (int, error) {
	conn := cach.GetConn()
	switch result := conn.(type) {
	case *redis.Pool:
		return hGetIntByRedis(result, key, field)
	case *ssdb.Client:
		return 0, nil
	default:
		return 0, nil
	}
}

func hGetIntByRedis(conn *redis.Pool, key string, field interface{}) (int, error) {
	c := conn.Get()
	defer c.Close()
	return redis.Int(c.Do("HGET", key, field))
}

func HgetString(cach cache.Cache, key string, field string) (string, error) {
	conn := cach.GetConn()
	switch result := conn.(type) {
	case *redis.Pool:
		return hGetStringByRedis(result, key, field)
	case *ssdb.Client:
		return "", nil
	default:
		return "", nil
	}
}

func hGetStringByRedis(conn *redis.Pool, key string, field string) (string, error) {
	c := conn.Get()
	defer c.Close()
	return redis.String(c.Do("HGET", key, field))
}

func Hexists(cach cache.Cache, key string, field string) (bool, error) {
	conn := cach.GetConn()
	switch result := conn.(type) {
	case *redis.Pool:
		return hExistsByRedis(result, key, field)
	case *ssdb.Client:
		return false, nil
	default:
		return false, nil
	}
}

func hExistsByRedis(conn *redis.Pool, key string, field string) (bool, error) {
	c := conn.Get()
	defer c.Close()
	return redis.Bool(c.Do("HEXISTS", key, field))
}

func Hlen(cach cache.Cache, key string) (int, error) {
	conn := cach.GetConn()
	switch result := conn.(type) {
	case *redis.Pool:
		return hLenByRedis(result, key)
	case *ssdb.Client:
		return 0, nil
	default:
		return 0, nil
	}
}

func hLenByRedis(conn *redis.Pool, key string) (int, error) {
	c := conn.Get()
	defer c.Close()
	return redis.Int(c.Do("HLEN", key))
}

func HmgetString(cach cache.Cache, keys ...interface{}) ([]string, error) {
	conn := cach.GetConn()
	switch result := conn.(type) {
	case *redis.Pool:
		return HmgetByRedis(result, keys...)
	case *ssdb.Client:
		return nil, nil
	default:
		return nil, nil
	}
}

func HmgetByRedis(conn *redis.Pool, keys ...interface{}) ([]string, error) {
	c := conn.Get()
	defer c.Close()
	var ret []string
	values, err := redis.Values(c.Do("HMGET", keys...))
	if err != nil {
		return nil, err
	}
	if err = redis.ScanSlice(values, &ret); err != nil {
		return nil, err
	}
	return ret, nil
}

func Hincrby(cach cache.Cache, key string, member interface{}, num int) error {
	conn := cach.GetConn()
	switch result := conn.(type) {
	case *redis.Pool:
		return hIncrbyByRedis(result, key, member, num)
	case *ssdb.Client:
		return nil
	default:
		return nil
	}
}

func hIncrbyByRedis(conn *redis.Pool, key string, field interface{}, num int) error {
	c := conn.Get()
	defer c.Close()
	_, err := c.Do("HINCRBY", key, field, num)
	return err
}

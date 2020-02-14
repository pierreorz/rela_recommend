package utils

import (
	"strings"
	"rela_recommend/cache"
	"rela_recommend/cache/redis"
	"rela_recommend/cache/redisCluster"
)

// 根据redis连接内是否有,分隔符。有分隔符使用cluster
func NewRedisOrClusterCache(conninfo, password string, dbNum int) (cache.Cache, error) {
	if strings.Contains(conninfo, ",") {
		return redisCluster.NewRedisCache(conninfo, password, dbNum)
	} else {
		return redis.NewRedisCache(conninfo, password, dbNum)
	}
}

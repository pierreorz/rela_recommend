package utils

import (
	"rela_recommend/cache"
	"rela_recommend/cache/redis"
	"strings"
)

// 根据redis连接内是否有,分隔符。有分隔符使用cluster
func NewRedisOrClusterCache(conninfo, password string, dbNum int) (cache.Cache, error) {
	if strings.Contains(conninfo, ",") {
		//TODO 有分号用集群
		return redis.NewRedisCache(conninfo, password, dbNum)
	} else {
		return redis.NewRedisCache(conninfo, password, dbNum)
	}
}

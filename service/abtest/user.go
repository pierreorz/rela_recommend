package abtest

import (
	"strings"
	"rela_recommend/rpc/api"
	// "rela_recommend/models/redis"
)

func GetUserAttr(userId int64, ua string, lat float32, lng float32, keys []string) map[string]interface{} {
	res := map[string]interface{}{
		"vip_level": 0, 							// 会员等级
		"os": "other", 								// 操作系统：ios,android,other
		// "registe_time": 60 * 60 * 24 * 365 * 100,	// 注册时间距离目前多少秒，默认10年前
		// "active_time": 60 * 60 * 24 * 365 * 100,	// 最后活跃时间距离目前多少秒，默认10年前
		// "version": "",								// 当前版本
	}
	if userId > 0 {
		for _, key := range keys {
			switch key {
				case "vip_level": {	// 会员等级
					vipRes, _ := api.CallUserVipStatusWithCache(userId, 1 * 60 * 60)
					res[key] = vipRes.Level
				}
				case "os": {
					if strings.Contains(ua, "iOS") {
						res[key] = "ios"
					} else if strings.Contains(ua, "android") {
						res[key] = "android"
					}
				}
				// case "registe_time": {	// 最后活跃时间， 注册时间
				// 	userCache := redis.NewUserCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)
				// }
			}
		}
	}
	return res
}

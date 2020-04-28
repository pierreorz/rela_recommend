package abtest

import (
	"strings"
	"rela_recommend/rpc/api"
	"rela_recommend/log"
	// "rela_recommend/models/redis"
)

func (self *AbTest)GetUserAttr(keys []string) map[string]interface{} {
	res := map[string]interface{}{}
	// "vip_level": 0, 							// 会员等级
	// "os": "other", 								// 操作系统：ios,android,other
	// "registe_time": 60 * 60 * 24 * 365 * 100,	// 注册时间距离目前多少秒，默认100年前
	// "active_time": 60 * 60 * 24 * 365 * 100,	// 最后活跃时间距离目前多少秒，默认100年前
	// "age": 0,									// 年龄
	
	if self.DataId > 0 {
		for _, key := range keys {
			switch key {
				case "vip_level": {	// 会员等级
					if vipRes, err := api.CallUserVipStatusWithCache(self.DataId, 1 * 60 * 60); err == nil {
						res[key] = vipRes.Level
					} else {
						res[key] = 0		// 默认0
						log.Warnf("abtest call user vip err %s", err)
					}
				}
				case "registe_time", "active_time", "age": {
					if userRes, err := api.CallUserInfoWithCache(self.DataId, 3 * 60 * 60); err == nil {
						res["active_time"] = self.CurrentTime.Unix() - userRes.LastUpdateTime
						res["registe_time"] = self.CurrentTime.Unix() - userRes.CreateTime
						res["age"] = userRes.Age
					} else {
						res["active_time"] = 60 * 60 * 24 * 365 * 100	// 默认100年前
						res["registe_time"] = 60 * 60 * 24 * 365 * 100	// 默认100年前
						res["age"] = 0		// 默认0
						log.Warnf("abtest call user info err %s", err)
					}
				}
				case "os": {
					lowerUa := strings.ToLower(self.Ua)
					if strings.Contains(self.Ua, "iOS") {
						res[key] = "ios"
					} else if strings.Contains(lowerUa, "android") {
						res[key] = "android"
					} else {
						res[key] = "other"
					}
				}
			}
		}
	}
	return res
}

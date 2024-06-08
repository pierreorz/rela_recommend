package abtest

import (
	"rela_recommend/log"
	"rela_recommend/rpc/api"
	"rela_recommend/utils"
	// "rela_recommend/models/redis"
)

func (self *AbTest) GetUserAttr(keys []string) map[string]interface{} {
	res := map[string]interface{}{}
	// "vip_level": 0, 							// 会员等级
	// "os": "other", 								// 操作系统：ios,android,other
	// "registe_time": 60 * 60 * 24 * 365 * 100,	// 注册时间距离目前多少秒，默认100年前
	// "active_time": 60 * 60 * 24 * 365 * 100,	// 最后活跃时间距离目前多少秒，默认100年前
	// "age": 0,									// 年龄

	if self.DataId > 0 {
		for _, key := range keys {
			switch key {
			case "vip_level": // 会员等级
				//if vipRes, err := api.CallUserVipStatusWithCache(self.DataId, 1*60*60); err == nil {
				//	res[key] = vipRes.Level
				//} else {
				res[key] = 0 // 默认0 TODO:改为读缓存
				//	log.Warnf("abtest call user vip err %s", err)
				//}
			case "registe_time", "active_time", "age":
				if userRes, err := api.CallUserInfoWithCache(self.DataId, 3*60*60); err == nil {
					res["active_time"] = self.CurrentTime.Unix() - userRes.LastUpdateTime
					res["registe_time"] = self.CurrentTime.Unix() - userRes.CreateTime
					res["age"] = userRes.Age
				} else {
					res["active_time"] = 60 * 60 * 24 * 365 * 100  // 默认100年前
					res["registe_time"] = 60 * 60 * 24 * 365 * 100 // 默认100年前
					res["age"] = 0                                 // 默认0
					log.Warnf("abtest call user info err %s", err)
				}
			case "os":
				res[key] = utils.GetPlatformName(self.Ua)
			case "client_version":
				res[key] = utils.GetVersion(self.Ua)
			case "os_type", "brand", "model_type", "net_type", "language":
				osType, brand, modelType, netType, language := utils.UaAnalysis(self.Ua)
				res["os_type"] = osType
				res["brand"] = brand
				res["model_type"] = modelType
				res["net_type"] = netType
				res["language"] = language
			case "lat", "lng":
				res["lat"] = self.Lat
				res["lng"] = self.Lng
			case "region":
				if self.SettingMap != nil {
					res["region"] = self.SettingMap["region"]
				} else {
					res["region"] = ""
				}
			}
		}
	}
	return res
}

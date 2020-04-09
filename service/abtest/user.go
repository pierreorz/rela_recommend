package abtest

import (
	"rela_recommend/rpc/api"
)

func GetUserAttr(userId int64, keys []string) map[string]interface{} {
	res := map[string]interface{}{}
	for _, key := range keys {
		switch key {
			case "vip_level": {	// 会员等级
				vipRes, _ := api.CallUserVipStatus(userId)
				res[key] = vipRes.Level
			}
		}
	}
	return res
}

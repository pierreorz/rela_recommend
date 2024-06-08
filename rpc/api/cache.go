package api

import (
	"fmt"
	"errors"
	"encoding/json"
	"rela_recommend/models/pika"
	"rela_recommend/factory"
)

func CallUserVipStatusWithCache(userId int64, cacheTime int) (userVipStatusDataRes, error) {
	key := fmt.Sprintf("rpc.user_vip_detail:%d", userId)
	var res userVipStatusDataRes
	var err error

	// 获取本地缓存
	var val interface{}
	if val, err = factory.CacheLoc.Get(key); err == nil {
		if val != nil {
			if valBs, ok := val.([]byte); ok {
				err = json.Unmarshal(valBs, &res)
			}
		} else {
			err = errors.New("local cache not found key")
		}
	}
	// 读取接口
	if err != nil {
		if res, err = CallUserVipStatus(userId); err == nil {
			if js, err := json.Marshal(res); err == nil {
				// 写入本地缓存
				factory.CacheLoc.SetEx(key, js, cacheTime)
			}
		}
	}
	return res, err
}



type userInfo struct {
	UserId         	int64    	`json:"id"`         // 用户ID
	IsVip          	int      	`json:"isVip"`          // 是否是vip
	LastUpdateTime 	int64    	`json:"lastUpdateTime"` //最后在线时间
	Age        		int       	`json:"age"`
	CreateTime 		int64 		`json:"createTime"`
}

func CallUserInfoWithCache(userId int64, cacheTime int) (userInfo, error) {
	key := fmt.Sprintf("cache.user_info:%d", userId)
	var res = userInfo{}
	// 获取本地缓存
	val, err := factory.CacheLoc.Get(key)
	if err != nil || val == nil {
		userCache := pika.NewUserProfileModule(&factory.CacheCluster, &factory.PikaCluster)
		if user, _, errCache := userCache.QueryByUserAndUsers(userId, []int64{}); errCache == nil {
			res.UserId = user.UserId
			res.IsVip = user.IsVip
			res.LastUpdateTime = user.LastUpdateTime
			res.Age = user.Age
			res.CreateTime = user.CreateTime.Time.Unix()

			if js, errJson := json.Marshal(res); errJson == nil {
				factory.CacheLoc.SetEx(key, js, cacheTime)	// 写入本地缓存
			}
		} else {
			err = errCache
		}
	} else {
		err = json.Unmarshal(val.([]byte), &res)
	}
	return res, err
}

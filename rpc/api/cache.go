package api

import (
	"fmt"
	"errors"
	"encoding/json"
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

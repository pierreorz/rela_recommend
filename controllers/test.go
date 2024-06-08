package controllers

import (
	"math/rand"
	"rela_recommend/factory"
	"rela_recommend/log"
	"rela_recommend/models/pika"
	"rela_recommend/service"
	"rela_recommend/utils/request"
	"rela_recommend/utils/response"
	"rela_recommend/utils/routers"
	"time"
)

type TestReqParams struct {
	Count int    `form:"count"`
	Type  string `form:"type"`
}

func GenUserIds(cnt int) []int64 {
	var maxId int64 = 105000000
	var minId int64 = 100000000
	userIds := make([]int64, 0, cnt)
	for i := 0; i < cnt; i++ {
		userIds = append(userIds, rand.Int63n(maxId-minId)+minId)
	}
	return userIds
}

func TestHTTP(c *routers.Context) {
	var startTime = time.Now()
	var params TestReqParams
	if err := request.Bind(c, &params); err != nil {
		log.Error(err.Error())
		c.JSON(response.FormatResponse(nil, service.WarpError(service.ErrInvaPara, "", "")))
		return
	}
	if params.Count <= 0 {
		params.Count = 3000
	}

	// 开始测试
	var resLen int
	userIds := GenUserIds(params.Count)
	if params.Type == "cache" { // 测试缓存
		aulm := pika.NewUserProfileModule(&factory.CacheCluster, &factory.PikaCluster)
		users, _ := aulm.QueryByUserIds(userIds)
		resLen = len(users)
	}
	var startLogTime = time.Now()
	log.Infof("userids:%d,total:%.3f", len(userIds), startLogTime.Sub(startTime).Seconds())
	c.JSON(response.FormatResponse(resLen, service.WarpError(nil, "", "")))
}

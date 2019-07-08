package controllers

import (
	"time"
	"rela_recommend/factory"
	"rela_recommend/models/pika"
	"rela_recommend/utils/routers"
	"rela_recommend/service"
	"math/rand"
	"rela_recommend/log"
	"rela_recommend/utils/response"
	"rela_recommend/utils/request"
	"rela_recommend/controllers/match"
)

type TestReqParams struct {
	Count   int     `form:"count"`
	Type    string  `form:"type"`
}


func GenUserIds(cnt int) []int64 {
	var maxId int64 = 105000000
	var minId int64 = 100000000
	userIds := make([]int64, 0, cnt)
	for i:=0; i < cnt; i++{
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
	var userId int64 = 104708381

	// 开始测试
	var resLen int
	userIds := GenUserIds(params.Count)
	if params.Type == "cache" {  // 测试缓存
		aulm := pika.NewUserProfileModule(&factory.CacheCluster, &factory.PikaCluster)
		users, _ := aulm.QueryByUserIds(userIds)
		resLen = len(users)
	} else if params.Type == "quick_match" {  // 测试速配
		pars := match.MatchRecommendReqParams{UserId: userId, Offset:0, Limit:20}
		res := match.DoRecommend(&pars, userIds)
		resLen = len(res.UserIds)
	}
	var startLogTime = time.Now()
	log.Infof("userids:%d,total:%.3f", len(userIds), startLogTime.Sub(startTime).Seconds())
	c.JSON(response.FormatResponse(resLen, service.WarpError(nil, "", "")))
}


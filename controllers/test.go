package controllers

import (
	"time"
	"rela_recommend/factory"
	"rela_recommend/models/mongo"
	"rela_recommend/routers"
	"rela_recommend/service"
	"rela_recommend/utils"
	"math/rand"
	"rela_recommend/log"
	"strings"
)

type TestReqParams struct {
	Limit   int64  `json:"limit" form:"limit"`
	Offset  int64  `json:"offset" form:"offset"`
	UserId  int64  `json:"userId" form:"userId"`
	UserIds string `json:"userIds" form:"userIds"`
}

func TestHTTP(c *routers.Context) {
	var startTime = time.Now()
	var params TestReqParams
	if err := bind(c, &params); err != nil {
		log.Error(err.Error())
		c.JSON(formatResponse(nil, service.WarpError(service.ErrInvaPara, "", "")))
		return
	}
	var userIds2 = make([]int64, 0)
	var userIds2Strs = strings.Split(params.UserIds, ",")
	for _, uid := range userIds2Strs {
		userIds2 = append(userIds2, utils.GetInt64(uid))
	}


	var mongoClient = factory.MatchClusterMon.Copy()
	defer mongoClient.Close()

	var maxId int64 = 105000000
	var minId int64 = 100000000
	userIds := make([]int64, 0)
	for i:=0;i<5000;i++{
		userIds = append(userIds, rand.Int63n(maxId-minId)+minId)
	}
	aulm := mongo.NewActiveUserLocationModule(mongoClient)
	aulm.QueryByUserIdsFromMongo(userIds)
	var startLogTime = time.Now()
	log.Infof("Test:userids:%d,total:%.3f", len(userIds), startLogTime.Sub(startTime).Seconds())
	
	res, err := aulm.QueryByUserIdsFromMongo(userIds2)
	c.JSON(formatResponse(res, service.WarpError(err, "", "")))
}


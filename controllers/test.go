package controllers

import (
	"time"
	"rela_recommend/factory"
	"rela_recommend/models/mongo"
	"rela_recommend/routers"
	"rela_recommend/service"
	"math/rand"
	"rela_recommend/log"
)


func TestHTTP(c *routers.Context) {
	var startTime = time.Now()
	// var params MatchRecommendReqParams

	var mongoClient = factory.MatchClusterMon.Copy()
	defer mongoClient.Close()

	var maxId int64 = 104887329
	var minId int64 = 104860000
	userIds := make([]int64, 0)
	for i:=0;i<3000;i++{
		userIds = append(userIds, rand.Int63n(maxId-minId)+minId)
	}
	aulm := mongo.NewActiveUserLocationModule(mongoClient)
	aulm.QueryByUserIdsFromMongo(userIds)
	var startLogTime = time.Now()
	log.Infof("Test:userids:%d,total:%.3f", len(userIds), startLogTime.Sub(startTime).Seconds())
	
	userIds2 := []int64 {110652}
	res, err := aulm.QueryByUserIdsFromMongo(userIds2)
	c.JSON(formatResponse(res, service.WarpError(err, "", "")))
}


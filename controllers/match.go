package controllers

import (
	"math"
	"rela_recommend/algo"
	"rela_recommend/algo/quick_match"
	"rela_recommend/factory"
	"rela_recommend/log"
	"rela_recommend/models/mongo"
	"rela_recommend/routers"
	"rela_recommend/service"
	"rela_recommend/utils"
	"sort"
	"strings"
)

type MatchRecommendReqParams struct {
	Limit   int64  `json:"limit" form:"limit"`
	Offset  int64  `json:"offset" form:"offset"`
	UserId  int64  `json:"userId" form:"userId"`
	UserIds string `json:"userIds" form:"userIds"`
}

type MatchRecommendResponse struct {
	Status  string
	RankId	string
	UserIds []int64
}

type MatchRecommendLog struct {
	RankId string
	Index int64
	UserId int64
	ReceiverId int64
	Score float32
	Features string
}

func MatchRecommendListHTTP(c *routers.Context) {
	var params MatchRecommendReqParams
	if err := bind(c, &params); err != nil {
		log.Error(err.Error())
		c.JSON(formatResponse(nil, service.WarpError(service.ErrInvaPara, "", "")))
		return
	}
	var userIds = make([]int64, 0)
	var userIdsStrs = strings.Split(params.UserIds, ",")
	for _, uid := range userIdsStrs {
		userIds = append(userIds, utils.GetInt64(uid))
	}

	var mongoClient = factory.MatchClusterMon.Copy()
	defer mongoClient.Close()

	rank_id := utils.UniqueId()
	// 加载用户缓存
	aulm := mongo.NewActiveUserLocationModule(mongoClient)
	// user, err1 := aulm.QueryOneByUserId(params.UserId)
	// if err1 != nil {
	// 	log.Error(err1.Error())
	// }
	// users, err2 := aulm.QueryByUserIds(userIds)
	// if err2 != nil {
	// 	log.Error(err2.Error())
	// }
	user, users, err := aulm.QueryByUserAndUsers(params.UserId, userIds)
	if err != nil {
		log.Error(err.Error())
	}
	userLen := len(users)
	// 构建上下文
	userInfo := &quick_match.UserInfo{UserId: user.UserId, UserCache: &user}
	usersInfo := make([]quick_match.UserInfo, userLen)
	for i, u := range users {
		usersInfo[i].UserId = u.UserId
		usersInfo[i].UserCache = &users[i]
	}
	ctx := quick_match.QuickMatchContext{RankId: rank_id, User: userInfo, UserList: usersInfo}
	// 算法预测打分
	quick_match.MatchAlgo.Predict(&ctx)
	// 结果排序
	sort.Sort(quick_match.UserInfoListSort(ctx.UserList))
	// 分页结果
	maxIndex := int64(math.Min(float64(len(ctx.UserList)), float64(params.Offset+params.Limit)))
	returnIds := make([]int64, maxIndex-params.Offset)
	for i := params.Offset; i < maxIndex; i++ {
		currUser := ctx.UserList[i]
		returnIds[i-params.Offset] = currUser.UserId
		// 记录日志
		logStr := MatchRecommendLog{RankId: rank_id,
									UserId: ctx.User.UserId,
									ReceiverId: currUser.UserId,
									Score: currUser.Score,
									Features: algo.Features2String(currUser.Features)}
		log.Infof("%+v\n", logStr)
	}
	log.Infof("param user %d,user %d,param len %d,len %d,return %d,max %g,min %g\n",
			  params.UserId, ctx.User.UserId, len(userIds), userLen, len(returnIds),
			  ctx.UserList[0].Score, ctx.UserList[userLen-1].Score)
	// 返回
	res := MatchRecommendResponse{RankId: rank_id, UserIds: returnIds, Status: "ok"}
	c.JSON(formatResponse(res, service.WarpError(nil, "", "")))
}

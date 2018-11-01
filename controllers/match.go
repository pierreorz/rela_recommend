package controllers
import (
	"sort"
	"math"
	"rela_recommend/routers"
	"rela_recommend/service"
	"rela_recommend/models/mongo"
	"rela_recommend/factory"
	"rela_recommend/log"
	"rela_recommend/algo"
	"rela_recommend/algo/quick_match"
)

type MatchRecommendReqParams struct {
	Limit int32 `json:"limit"`
	Offset int32 `json:"offset"`
	UserId int64 `json:"userId"`
	UserIds string `json:"userIds"`
}

func MatchRecommendListHTTP(c *routers.Context) {
	var params MatchRecommendReqParams
	if err := bind(c, &params); err != nil {
		log.Error(err.Error())
		c.JSON(formatResponse(nil, service.WarpError(service.ErrInvaPara, "", "")))
		return
	}

	// 加载用户缓存, 构建上下文
	aulm := mongo.NewActiveUserLocationModule(factory.MatchClusterMon)
	user, oerr := &aulm.QueryOneByUserId(params.UserId)
	users, merr := &aulm.QueryByUserIds(params.UserIds)
	userInfo := algo.UserInfo{UserId=user.UserId, UserCache=user}
	usersInfo := make([]algo.UserInfo, len(users))
	for i, u := range users {
		usersInfo[i].UserId = u.UserId
		usersInfo[i].UserCache = &u
	}
	ctx := quick_match.QuickMatchContext{User=&userInfo, UserList=&usersInfo}
	// 算法预测打分
	matchAlgo.Predict(&ctx)
	// 结果排序
	sort.Reverse(algo.UserInfoSortReverse(ctx.usersInfo))
	// 返回结果
	maxIndex := math.Min(len(ctx.usersInfo), params.Offset + params.Limit)
	returnIds := make([]int64, maxIndex - params.Offset)
	for i:=params.Offset; i<=maxIndex; i++ {
		returnIds[i-params.Offset] =  ctx.usersInfo[i].UserId
	}
	data["userIds"] = returnIds
	data["status"] = "ok"
	c.JSON(formatResponse(data, service.WarpError(nil, "", "")))
}
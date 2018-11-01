package controllers
import (
	"sort"
	"math"
	"rela_recommend/routers"
	"rela_recommend/service"
	"rela_recommend/models/mongo"
	"rela_recommend/factory"
	"rela_recommend/log"
	"rela_recommend/algo/quick_match"
)

type MatchRecommendReqParams struct {
	Limit int64 `json:"limit"`
	Offset int64 `json:"offset"`
	UserId int64 `json:"userId"`
	UserIds []int64 `json:"userIds"`
}

type MatchRecommendResponse struct {
	Status string
	UserIds []int64
}

func MatchRecommendListHTTP(c *routers.Context) {
	var params MatchRecommendReqParams
	if err := bind(c, &params); err != nil {
		log.Error(err.Error())
		c.JSON(formatResponse(nil, service.WarpError(service.ErrInvaPara, "", "")))
		return
	}

	// 加载用户缓存
	aulm := mongo.NewActiveUserLocationModule(factory.MatchClusterMon)
	user, _ := aulm.QueryOneByUserId(params.UserId)
	users, _ := aulm.QueryByUserIds(params.UserIds)
	// 构建上下文
	userInfo := &quick_match.UserInfo{UserId: user.UserId, UserCache: &user}
	usersInfo := make([]quick_match.UserInfo, len(users))
	for i, u := range users {
		usersInfo[i].UserId = u.UserId
		usersInfo[i].UserCache = &u
	}
	ctx := quick_match.QuickMatchContext{User: userInfo, UserList: usersInfo}
	// 算法预测打分
	quick_match.MatchAlgo.Predict(&ctx)
	// 结果排序
	sr := sort.Reverse(quick_match.UserInfoSortReverse(ctx.UserList))
	sort.Sort(sr)
	// 返回结果
	maxIndex := int64(math.Min(float64(len(ctx.UserList)), float64(params.Offset + params.Limit)))
	returnIds := make([]int64, maxIndex - params.Offset)
	for i:=params.Offset; i <= maxIndex; i++ {
		returnIds[i-params.Offset] =  ctx.UserList[i].UserId
	}
	res := MatchRecommendResponse{UserIds: returnIds, Status: "ok"}
	c.JSON(formatResponse(res, service.WarpError(nil, "", "")))
}
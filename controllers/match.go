package controllers

import (
	"fmt"
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
	UserIds []int64
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
	fmt.Println("params users len:", len(userIds))

	// 加载用户缓存
	aulm := mongo.NewActiveUserLocationModule(factory.MatchClusterMon)
	user, _ := aulm.QueryOneByUserId(params.UserId)
	users, _ := aulm.QueryByUserIds(userIds)
	fmt.Println("cache users len:", len(users))
	// 构建上下文
	userInfo := &quick_match.UserInfo{UserId: user.UserId, UserCache: &user}
	usersInfo := make([]quick_match.UserInfo, len(users))
	for i, u := range users {
		usersInfo[i].UserId = u.UserId
		usersInfo[i].UserCache = &u
	}
	ctx := quick_match.QuickMatchContext{User: userInfo, UserList: usersInfo}
	fmt.Println("ctx users len:", len(ctx.UserList))
	// 算法预测打分
	quick_match.MatchAlgo.Predict(&ctx)
	// 结果排序
	sr := sort.Reverse(quick_match.UserInfoSortReverse(ctx.UserList))
	sort.Sort(sr)
	fmt.Println("sort users len:", len(ctx.UserList))
	// 分页结果
	maxIndex := int64(math.Min(float64(len(ctx.UserList)), float64(params.Offset+params.Limit)))
	returnIds := make([]int64, maxIndex-params.Offset)
	for i := params.Offset; i < maxIndex; i++ {
		currUser := ctx.UserList[i]
		returnIds[i-params.Offset] = currUser.UserId
		fmt.Println(ctx.User.UserId, currUser.UserId, currUser.Score, algo.Features2String(currUser.Features))
	}
	fmt.Println("return users len:", len(returnIds))
	// 返回
	res := MatchRecommendResponse{UserIds: returnIds, Status: "ok"}
	c.JSON(formatResponse(res, service.WarpError(nil, "", "")))
}

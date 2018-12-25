package match

import (
	"math"
	"time"
	"rela_recommend/algo"
	"rela_recommend/algo/quick_match"
	"rela_recommend/factory"
	"rela_recommend/log"
	"rela_recommend/models/pika"
	"rela_recommend/routers"
	"rela_recommend/service"
	"rela_recommend/utils"
	"sort"
	"strings"
	"rela_recommend/utils/response"
	"rela_recommend/utils/request"
)

type MatchRecommendReqParams struct {
	Limit   int64  `json:"limit" form:"limit"`
	Offset  int64  `json:"offset" form:"offset"`
	Ua      string `json:"ua" form:"ua"`
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
	Algo string
	AlgoScore float32
	Score float32
	Features string
}

func MatchRecommendListHTTP(c *routers.Context) {
	var params MatchRecommendReqParams
	if err := request.Bind(c, &params); err != nil {
		log.Error(err.Error())
		c.JSON(response.FormatResponse(nil, service.WarpError(service.ErrInvaPara, "", "")))
		return
	}
	var userIds = make([]int64, 0)
	var userIdsStrs = strings.Split(params.UserIds, ",")
	for _, uid := range userIdsStrs {
		userIds = append(userIds, utils.GetInt64(uid))
	}
	res := DoRecommend(&params, userIds)
	c.JSON(response.FormatResponse(res, service.WarpError(nil, "", "")))
}

func DoRecommend(params *MatchRecommendReqParams, userIds []int64) MatchRecommendResponse {
	var startTime = time.Now()
	rank_id := utils.UniqueId()
	// 加载用户缓存
	var startCacheTime = time.Now()
	aulm := pika.NewUserProfileModule(&factory.CacheCluster, &factory.PikaCluster)
	user, users, err := aulm.QueryByUserAndUsers(params.UserId, userIds)
	if err != nil {
		log.Error(err.Error())
	}
	userLen := len(users)
	// 构建上下文
	var startCtxTime = time.Now()
	userInfo := &quick_match.UserInfo{UserId: user.UserId, UserCache: &user}
	usersInfo := make([]quick_match.UserInfo, userLen)
	for i, u := range users {
		usersInfo[i].UserId = u.UserId
		usersInfo[i].UserCache = &users[i]
	}
	ctx := quick_match.QuickMatchContext{
		RankId: rank_id, Ua: params.Ua,
		User: userInfo, UserList: usersInfo}
	// 算法预测打分
	var startPredictTime = time.Now()

	var model quick_match.IQuickMatch = &quick_match.MatchAlgoV1_0
	if (ctx.User.UserId % 100 < 10) {
		model = &quick_match.MatchAlgoV1_2
	} else if (ctx.User.UserId % 100 < 40) {
		model = &quick_match.MatchAlgoV1_3
	}
	model.Predict(&ctx)
	// 提升活跃用户权重
	Active24HourUpper(&ctx)
	// 结果排序
	var startSortTime = time.Now()
	sort.Sort(quick_match.UserInfoListSort(ctx.UserList))
	// 分页结果
	var startPageTime = time.Now()
	maxIndex := int64(math.Min(float64(len(ctx.UserList)), float64(params.Offset+params.Limit)))
	returnIds := make([]int64, maxIndex-params.Offset)
	for i := params.Offset; i < maxIndex; i++ {
		j := i - params.Offset
		currUser := ctx.UserList[i]
		returnIds[j] = currUser.UserId
		// 记录日志
		logStr := MatchRecommendLog{RankId: rank_id, Index: j,
									UserId: ctx.User.UserId,
									ReceiverId: currUser.UserId,
									Algo: model.Name(),
									AlgoScore: currUser.AlgoScore,
									Score: currUser.Score,
									Features: algo.Features2String(currUser.Features)}
		log.Infof("%+v\n", logStr)
	}
	var startLogTime = time.Now()
	log.Infof("paramuser %d,user %d,paramlen %d,len %d,return %d,max %g,min %g;total:%.3f,init:%.3f,cache:%.3f,ctx:%.3f,predict:%.3f,sort:%.3f,page:%.3f\n",
			  params.UserId, ctx.User.UserId, len(userIds), userLen, len(returnIds),
			  ctx.UserList[0].Score, ctx.UserList[userLen-1].Score,
			  startLogTime.Sub(startTime).Seconds(), startCacheTime.Sub(startTime).Seconds(),
			  startCtxTime.Sub(startCacheTime).Seconds(), startPredictTime.Sub(startCtxTime).Seconds(),
			  startSortTime.Sub(startPredictTime).Seconds(), startPageTime.Sub(startSortTime).Seconds(),
			  startLogTime.Sub(startPageTime).Seconds())
	// 返回
	res := MatchRecommendResponse{RankId: rank_id, UserIds: returnIds, Status: "ok"}
	return res
}

func Active24HourUpper(ctx *quick_match.QuickMatchContext) {
	before24HourTime := time.Now().Unix() - 24 * 60 * 60
	for i, user := range ctx.UserList {
		if user.UserCache.LastUpdateTime >= before24HourTime {
			ctx.UserList[i].Score = ctx.UserList[i].Score * 1.1
		}
	}
}
package moment

import (
	"time"
	"rela_recommend/algo"
	"rela_recommend/log"
	"rela_recommend/factory"
	"rela_recommend/rpc/search"
	"rela_recommend/routers"
	"rela_recommend/algo/moment"
	"rela_recommend/service"
	"rela_recommend/models/pika"
	"rela_recommend/models/redis"
	"rela_recommend/utils"
	"rela_recommend/utils/response"
	"rela_recommend/utils/request"
)

func RecommendListHTTP(c *routers.Context) {
	var params = &algo.RecommendRequest{}
	if err := request.Bind(c, params); err != nil {
		log.Error(err.Error())
		c.JSON(response.FormatResponse(nil, service.WarpError(service.ErrInvaPara, "", "")))
		return
	}

	app := &algo.AppInfo{
		Name: "moment",
		AlgoKey: "model", AlgoMap: moment.AlgosMap,
		StrategyKey: "strategies", StrategyMap: moment.StrategyMap,
		SorterKey: "sorter", SorterMap: moment.SorterMap,
		PagerKey: "pager", PagerMap: moment.PagerMap,
		LoggerKey: "loggers", LoggerMap: moment.LoggerMap}
	ctx := &algo.ContextBase{}
	err := ctx.Do(app, params, DoBuildData)
	c.JSON(response.FormatResponse(ctx.GetResponse(), service.WarpError(err, "", "")))
}

func DoBuildData(ctx algo.IContext) error {
	var err error
	var startTime = time.Now()
	params := ctx.GetRequest()
	userCache := pika.NewUserProfileModule(&factory.CacheCluster, &factory.PikaCluster)
	momentCache := redis.NewMomentCacheModule(&factory.CacheCluster, &factory.PikaCluster)

	// search list
	dataIds := params.DataIds
	if dataIds == nil || len(dataIds) == 0 {
		dataIds, err = search.CallNearMomentList(params.UserId, params.Lat, params.Lng, 0, 1000)
		if err != nil {
			return err
		}
	}

	// 获取日志内容
	var startMomentTime = time.Now()
	moms, err := momentCache.QueryMomentsByIds(dataIds)
	userIds := make([]int64, 0)
	if err != nil {
		log.Warnf("moment list is err, %s\n", err)
	} else {
		for _, mom := range moms {
			if mom.Moments != nil {
				userIds = append(userIds, mom.Moments.UserId)
			}
		}
		userIds = utils.NewSetInt64FromArray(userIds).ToList()
	}
	// 获取用户信息
	var startUserTime = time.Now()
	user, usersMap, err := userCache.QueryByUserAndUsersMap(params.UserId, userIds)
	if err != nil {
		log.Warnf("users list is err, %s\n", err)
	}

	var startBuildTime = time.Now()
	userInfo := &moment.UserInfo{
		UserId: params.UserId,
		UserCache: user}

	dataList := make([]algo.IDataInfo, 0)
	for _, mom := range moms {
		if mom.Moments != nil && mom.Moments.Id > 0 {
			momUser, _ := usersMap[mom.Moments.UserId]
			info := &moment.DataInfo{
				DataId: mom.Moments.Id,
				UserCache: momUser,
				MomentCache: mom.Moments,
				MomentExtendCache: mom.MomentsExtend,
				MomentProfile: mom.MomentsProfile,
				RankInfo: &algo.RankInfo{}}
			dataList = append(dataList, info)
		}
	}
	ctx.SetUserInfo(userInfo)
	ctx.SetDataList(dataList)
	var endTime = time.Now()
	log.Infof("rankid %s,searchlen:%d;total:%.3f,search:%.3f,moment:%.3f,user:%.3f,build:%.3f\n",
			  ctx.GetRankId(), len(dataIds),
			  endTime.Sub(startTime).Seconds(), startMomentTime.Sub(startTime).Seconds(),
			  startUserTime.Sub(startMomentTime).Seconds(), startBuildTime.Sub(startUserTime).Seconds(),
			  endTime.Sub(startBuildTime).Seconds() )
	return nil
}

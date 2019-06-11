package moment

import (
	"time"
	"rela_recommend/algo"
	"rela_recommend/log"
	"rela_recommend/factory"
	"rela_recommend/routers"
	"rela_recommend/algo/moment"
	"rela_recommend/service"
	"rela_recommend/models/pika"
	"rela_recommend/models/redis"
	"rela_recommend/utils"
	"rela_recommend/utils/response"
	"rela_recommend/utils/request"
)

func CoarseRecommendListHTTP(c *routers.Context) {
	var params = &algo.RecommendRequest{}
	if err := request.Bind(c, params); err != nil {
		log.Error(err.Error())
		c.JSON(response.FormatResponse(nil, service.WarpError(service.ErrInvaPara, "", "")))
		return
	}

	app := &algo.AppInfo{
		Name: "moment_coarse",
		AlgoKey: "model", AlgoMap: moment.AlgosCoarseMap,
		StrategyKey: "strategies", StrategyMap: nil,
		SorterKey: "sorter", SorterMap: nil,
		PagerKey: "pager", PagerMap: nil,
		LoggerKey: "loggers", LoggerMap: nil}
	ctx := &algo.ContextBase{}
	err := ctx.Do(app, params, DoBuildCoarseData)
	c.JSON(response.FormatResponse(ctx.GetResponse(), service.WarpError(err, "", "")))
}

// 构建上下文
func DoBuildCoarseData(ctx algo.IContext) error {
	var startTime = time.Now()
	params := ctx.GetRequest()
	userCache := pika.NewUserProfileModule(&factory.CacheCluster, &factory.PikaCluster)
	momentCache := redis.NewMomentCacheModule(&factory.CacheCluster, &factory.PikaCluster)

	// search list
	var err error
	dataIds := params.DataIds
	// 获取日志内容
	var startMomentTime = time.Now()
	moms, err := momentCache.QueryMomentsByIds(dataIds)
	userIds := make([]int64, 0)
	if err != nil {
		return err
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
	usersMap, err := userCache.QueryUsersMap(userIds)

	var startBuildTime = time.Now()
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

	ctx.SetDataIds(dataIds)
	ctx.SetDataList(dataList)

	var endTime = time.Now()
	log.Infof("rankid %s,searchlen:%d;total:%.3f,other:%.3f,moment:%.3f,user:%.3f,build:%.3f\n",
			  ctx.GetRankId(), len(dataIds),
			  endTime.Sub(startTime).Seconds(), startMomentTime.Sub(startTime).Seconds(),
			  startUserTime.Sub(startMomentTime).Seconds(), startBuildTime.Sub(startUserTime).Seconds(),
			  endTime.Sub(startBuildTime).Seconds() )
	return nil
}

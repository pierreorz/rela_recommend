package ad

import (
	"rela_recommend/algo"
	"rela_recommend/rpc/search"
	"rela_recommend/factory"
	"rela_recommend/models/redis"
)

func DoBuildData(ctx algo.IContext) error {
	var err error
	abtest := ctx.GetAbTest()
	pf := ctx.GetPerforms()
	params := ctx.GetRequest()
	userCache := redis.NewUserCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)
	// behaviorCache := behavior.NewBehaviorCacheModule(ctx, &factory.CacheBehaviorRds)
	
	// 获取用户信息
	pf.Begin("user")
	user, _, userCacheErr := userCache.QueryByUserAndUsersMap(params.UserId, []int64{})
	if userCacheErr != nil {

	}
	pf.End("user")

	// 获取search的广告列表
	pf.Begin("search")
	clientName := abtest.GetString("app_name", "rela")
	searchResList, errSearch := search.CallAdList(clientName, params, user)
	if errSearch != nil {

	}
	pf.End("search")

	pf.Begin("build")
	userInfo := &UserInfo{
		UserId: params.UserId,
		UserCache: user,
	}

	// 组装被曝光者信息
	dataIds := make([]int64, 0)
	dataList := make([]algo.IDataInfo, 0)
	for i, searchRes := range searchResList {
		info := &DataInfo{
			DataId: searchRes.Id,
			SearchData: &searchResList[i],
			RankInfo: &algo.RankInfo{},
		}
		dataIds = append(dataIds, searchRes.Id)
		dataList = append(dataList, info)
	}
	ctx.SetUserInfo(userInfo)
	ctx.SetDataIds(dataIds)
	ctx.SetDataList(dataList)
	pf.End("build")

	return err
}
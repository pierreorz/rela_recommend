package theme

import (
	"rela_recommend/algo"
	"rela_recommend/log"
	"rela_recommend/factory"
	"rela_recommend/routers"
	"rela_recommend/algo/theme"
	"rela_recommend/service"
	// "rela_recommend/models/pika"
	"rela_recommend/models/redis"
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
		Name: "theme",
		AlgoKey: "model", AlgoDefault: "model_base", AlgoMap: nil,
		StrategyKey: "strategies", StrategyDefault: "time_level", StrategyMap: nil,
		SorterKey: "sorter", SorterDefault: "base", SorterMap: nil,
		PagerKey: "pager", PagerDefault: "base", PagerMap: theme.PagerMap,
		LoggerKey: "loggers", LoggerDefault: "features,performs", LoggerMap: theme.LoggerMap}
	ctx := &algo.ContextBase{}
	err := ctx.Do(app, params, DoBuildData)
	c.JSON(response.FormatResponse(ctx.GetResponse(), service.WarpError(err, "", "")))
}

func DoBuildData(ctx algo.IContext) error {
	params := ctx.GetRequest()
	rdsPikaCache := redis.NewUserProfileModule(&factory.CacheCluster, &factory.PikaCluster)
	dataIds, err := rdsPikaCache.GetInt64List(params.UserId, "theme_recommend_list:%d")
	if err == nil {
		log.Warnf("theme recommend list is nil, %s\n", err)
	}
	if len(dataIds) == 0{
		dataIds, _ = rdsPikaCache.GetInt64List(-999999999, "theme_recommend_list:%d")
	}
	user := &theme.UserInfo{UserId: params.UserId}
	dataList := []algo.IDataInfo{}
	for _, dataId := range dataIds {
		dataInfo := &theme.DataInfo{
			DataId: dataId,
			RankInfo: &algo.RankInfo{} }
		dataList = append(dataList, dataInfo)
	}
	ctx.SetUserInfo(user)
	ctx.SetDataIds(dataIds)
	ctx.SetDataList(dataList)
	return nil
}

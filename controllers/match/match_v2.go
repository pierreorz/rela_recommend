package match

import (
	"rela_recommend/algo"
	"rela_recommend/algo/base"
	"rela_recommend/algo/match"
	"rela_recommend/factory"
	"rela_recommend/log"
	"rela_recommend/models/pika"
	"rela_recommend/utils/routers"
	"rela_recommend/service"
	"rela_recommend/utils/response"
	"rela_recommend/utils/request"
)

func MatchRecommendListV2HTTP(c *routers.Context) {
	var params = &algo.RecommendRequest{}
	if err := request.Bind(c, params); err != nil {
		log.Error(err.Error())
		c.JSON(response.FormatResponse(nil, service.WarpError(service.ErrInvaPara, "", "")))
		return
	}
	app := &algo.AppInfo{
		Name: "match",
		AlgoKey: "model", AlgoMap: match.MatchAlgosMap,
		SorterKey: "sorter", SorterMap: nil,
		PagerKey: "pager", PagerMap: nil,
		StrategyKeyFormatter: "strategy:%s:weight", StrategyMap: nil,
		LoggerMap: nil}
	ctx := &base.ContextBase{}
	err := ctx.Do(app, params)
	c.JSON(response.FormatResponse(ctx.GetResponse(), service.WarpError(err, "", "")))
}

func DoBuildData(ctx algo.IContext) error {
	params := ctx.GetRequest()
	userCache := pika.NewUserProfileModule(&factory.CacheCluster, &factory.PikaCluster)
	user, usersMap, err := userCache.QueryByUserAndUsersMap(params.UserId, params.DataIds)
	if err != nil {
		return err	
	}
	userInfo := &match.UserInfo{ UserId: params.UserId, UserCache: user }
	dataList := []algo.IDataInfo{}
	for id, iuser := range usersMap {
		if iuser != nil && iuser.UserId > 0 {
			info := &match.DataInfo{DataId: id, UserCache: iuser, 
									RankInfo: &algo.RankInfo{} }
			dataList = append(dataList, info)
		}
	}

	ctx.SetDataIds(params.DataIds)
	ctx.SetUserInfo(userInfo)
	ctx.SetDataList(dataList)
	return nil
}
package theme

import (
	"rela_recommend/algo"
	"rela_recommend/log"
	"rela_recommend/factory"
	// "rela_recommend/models/pika"
	"rela_recommend/models/redis"
)

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
	user := &UserInfo{UserId: params.UserId}
	dataList := []algo.IDataInfo{}
	for _, dataId := range dataIds {
		dataInfo := &DataInfo{
			DataId: dataId,
			RankInfo: &algo.RankInfo{} }
		dataList = append(dataList, dataInfo)
	}
	ctx.SetUserInfo(user)
	ctx.SetDataIds(dataIds)
	ctx.SetDataList(dataList)
	return nil
}

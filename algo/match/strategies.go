package match

import (
	"time"
	"rela_recommend/algo"
)


// 对24小时内活跃用户进行提权
func ActiveUserUpperItem(ctx algo.IContext, iDataInfo algo.IDataInfo, rankInfo *algo.RankInfo) error {
	var offsetTime int64 = 1 * 60 * 60
	nowTime := time.Now().Unix()
	before24HourTime := nowTime - offsetTime

	dataInfo := iDataInfo.(*DataInfo)

	after24hourTime := dataInfo.UserCache.LastUpdateTime - before24HourTime
	if after24hourTime >= 0 {
		upperRate := ctx.GetAbTest().GetFloat("match_active_user_upper", 0.1)

		var addRate = float32(after24hourTime) / float32(offsetTime) * upperRate
		rankInfo.AddRecommend("ActiveUserUpper", addRate)
	}
	return nil
}

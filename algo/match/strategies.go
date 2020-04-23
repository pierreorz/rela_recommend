package match

import (
	"rela_recommend/algo"
	"time"
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

// 对有头像的用户进行提权
func ImageFaceUpperItem(ctx algo.IContext, iDataInfo algo.IDataInfo, rankInfo *algo.RankInfo) error {
	dataInfo := iDataInfo.(*DataInfo)
	currMatch := dataInfo.MatchProfile

	hasCover := currMatch.ImageMap["has_cover"]
	coverHasFace := currMatch.ImageMap["cover_has_face"]
	countImageWall := currMatch.ImageMap["imagewall_count"]
	wallHasFace := currMatch.ImageMap["imagewall_has_face"]
	headHasFace := currMatch.ImageMap["head_has_face"]

	upperRate := ctx.GetAbTest().GetFloat("match_face_upper", 0)

	if (coverHasFace == 1 && hasCover == 1) || (wallHasFace == 1 && countImageWall > 0) || headHasFace == 1 {
		rankInfo.AddRecommend("ImageFaceUpper", 1.0+upperRate)
	}
	return nil
}

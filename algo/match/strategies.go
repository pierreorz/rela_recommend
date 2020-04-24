package match

import (
	"rela_recommend/algo"
	rutils "rela_recommend/utils"
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

	if currMatch != nil {
		hasCover := rutils.GetInt(currMatch.ImageMap["has_cover"])
		coverHasFace := rutils.GetInt(currMatch.ImageMap["cover_has_face"])
		countImageWall := rutils.GetInt(currMatch.ImageMap["imagewall_count"])
		wallHasFace := rutils.GetInt(currMatch.ImageMap["imagewall_has_face"])
		headHasFace := rutils.GetInt(currMatch.ImageMap["head_has_face"])

		upperRate := ctx.GetAbTest().GetFloat("match_face_upper", 0)

		coverFace := coverHasFace == 1 && hasCover == 1
		wallFace := wallHasFace == 1 && countImageWall > 0
		headFace := headHasFace == 1

		if coverFace || wallFace || headFace {
			rankInfo.AddRecommend("ImageFaceUpper", 1.0+upperRate)
		}
	}
	return nil
}

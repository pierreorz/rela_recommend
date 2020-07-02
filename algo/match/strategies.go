package match

import (
	"rela_recommend/algo"
	"rela_recommend/log"
	"rela_recommend/rpc/search"
	"time"
)

// 对在线活跃用户进行提权
func ActiveUserUpperItem(ctx algo.IContext, iDataInfo algo.IDataInfo, rankInfo *algo.RankInfo) error {
	var offsetTime int64 = 0.5 * 60 * 60
	nowTime := time.Now().Unix()
	beforeTime := nowTime - offsetTime

	dataInfo := iDataInfo.(*DataInfo)

	afterTime := dataInfo.UserCache.LastUpdateTime - beforeTime
	if afterTime >= 0 {
		upperRate := ctx.GetAbTest().GetFloat("match_active_user_upper", 0.1)

		var addRate = float32(afterTime) / float32(offsetTime) * upperRate
		rankInfo.AddRecommend("ActiveUserUpper", 1.0+addRate)
	}
	return nil
}

// 对有头像的用户进行提权
func ImageFaceUpperItem(ctx algo.IContext, iDataInfo algo.IDataInfo, rankInfo *algo.RankInfo) error {
	dataInfo := iDataInfo.(*DataInfo)
	currMatch := dataInfo.MatchProfile

	if currMatch != nil {
		hasCover := currMatch.ImageMap["has_cover"] == 1
		coverHasFace := currMatch.ImageMap["cover_has_face"] == 1
		countImageWall := currMatch.ImageMap["imagewall_count"]
		wallHasFace := currMatch.ImageMap["imagewall_has_face"] == 1
		headHasFace := currMatch.ImageMap["head_has_face"] == 1

		upperRate := ctx.GetAbTest().GetFloat("match_face_upper", 0)

		//是否有脸
		hasFace := false

		// 是否有封面
		if hasCover {
			// 封面是否有脸
			if coverHasFace {
				hasFace = true
			}
			// 是否有照片墙
		} else if countImageWall > 0 {
			// 照片墙是否有脸
			if wallHasFace {
				hasFace = true
			}
		} else {
			// 头像是否有脸
			if headHasFace {
				hasFace = true
			}
		}
		if hasFace {
			rankInfo.AddRecommend("ImageFaceUpper", 1.0+upperRate)
		}
	}
	return nil
}

type DoMatchSeenSearchLogger struct{}

// 已读接口调用
func (self *DoMatchSeenSearchLogger) Do(ctx algo.IContext) error {
	response := ctx.GetResponse()
	seenIds := make([]int64, 0)
	if response != nil {
		for _, item := range response.DataList {
			seenIds = append(seenIds, item.DataId)
		}
	}
	if len(seenIds) > 0 {
		go func() {
			ok := search.CallMatchSeenList(ctx.GetRequest().UserId, 60*60, "", seenIds)
			if !ok {
				log.Warn("search seen failed\n")
			}
		}()
	}
	return nil
}

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

// 对已婚、交往中等降权
// affection  感情状态	int	‘-1表示未设置 0=不想透漏 1=单身 2=约会中 3=稳定关系 4=已婚 5=开放关系 6=求交往 7=等一个人
func NotSingleDecreaseItem(ctx algo.IContext, iDataInfo algo.IDataInfo, rankInfo *algo.RankInfo) error {
	notSingleStatus := []int{2, 3, 4}
	matchUser := iDataInfo.(*DataInfo)

	if matchUser.UserCache != nil {
		for _, st := range notSingleStatus {
			if matchUser.UserCache.Affection == st {
				rankInfo.AddRecommendNeedReturn("NotSingleDecrease", 0.5)
			}
		}
	}

	return nil
}

// 参考数据分析：https://das.base.shuju.aliyun.com/product/view.htm?productId=2242a5a3afe249feb0ecebb8b71f8fc7&menuId=tsh720r7at8
// 新注册用户营收转化率远远高于总体用户
func NewUserUpperItem(ctx algo.IContext, iDataInfo algo.IDataInfo, rankInfo *algo.RankInfo) error {
	matchUser := iDataInfo.(*DataInfo)

	if matchUser.UserCache != nil {
		if ctx.GetCreateTime().Sub(matchUser.UserCache.CreateTime.Time) <= time.Hour*24*7 {
			rankInfo.AddRecommend("NewUserUpper", 1.1)
		}
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

// 对有头像的用户进行提权
func ImageFaceUpperItemV2(ctx algo.IContext, iDataInfo algo.IDataInfo, rankInfo *algo.RankInfo) error {
	matchUser := iDataInfo.(*DataInfo)
	if (matchUser.SearchFields != nil) && (matchUser.SearchFields.CoverHasFace) {
		upperRate := ctx.GetAbTest().GetFloat("match_face_upper", 0)
		rankInfo.AddRecommend("ImageFaceUpperV2", upperRate)
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

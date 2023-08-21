package mate

import (
	"math/rand"
	"rela_recommend/algo"
	"rela_recommend/utils"
	"strings"
)

// 内容较短，包含关键词的内容沉底
func BaseScoreStrategyItem(ctx algo.IContext, iDataInfo algo.IDataInfo, rankInfo *algo.RankInfo) error {
	abtest := ctx.GetAbTest()
	dataInfo := iDataInfo.(*DataInfo)
	sd := dataInfo.SearchData

	abSwitch := abtest.GetBool("mate_text_switch", false)
	if abSwitch {
		randomScore := float32(rand.Intn(100)) / 100.0
		rankInfo.Score = float32(sd.Weight) + randomScore
	}

	return nil
}
//多种类型的分发策略
func SortScoreItem(ctx algo.IContext) error {
	//var itemWeightMap= make(map[int64]int)
	abtest := ctx.GetAbTest()
	if abtest.GetBool("meta_list_result",false) {
		//后台配置曝光权重
		admin_weight := abtest.GetStrings("sentence_type_weight", "10:1,20:1.05,30:1,40:1,50:1,60:1.05,70:1.05")
		adminMap := make(map[int64]float64)
		for _, backtag := range admin_weight {
			type_nums := utils.GetInt64(strings.Split(backtag, ":")[0])
			admin_weight_num := utils.GetFloat64(strings.Split(backtag, ":")[1])
			adminMap[type_nums] = admin_weight_num
		}
		//曝光逻辑
		var sdWeight float64
		var itemScore float32
		for index := 0; index < ctx.GetDataLength(); index++ {
			//随机对文案给出一个权重，
			randomScore := float32(rand.Intn(100)) / 100.0
			dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
			sd := dataInfo.SearchData //SearchData缺少类型信息,
			rankInfo := dataInfo.GetRankInfo()
			//sdWeight为 管理后台可配置权重
			if _, ok := adminMap[sd.TextType]; ok {
				sdWeight = adminMap[sd.TextType]
			} else {
				sdWeight = 1.0
			}
			//对于以前的文案全部变成兜底设置
			if sd.UserId == 0 {
				itemScore = float32(0.0001)
			} else {
				itemScore = randomScore * float32(sdWeight) * float32(sd.Weight)
			}
			//log.Infof("mate_text=====itemScore",sd.Id,sd.Text,itemScore)
			rankInfo.AddRecommend("sortScoreItem", itemScore)
		}
		return nil
	}
	return nil
}


package mate

import (
	"math/rand"
	"rela_recommend/algo"
	"rela_recommend/log"
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
	//后台配置曝光权重
	admin_weight := abtest.GetStrings("sentence_type_weight", "10:1,20:1,30:1,40:1,50:1")
	adminMap := make(map[int64]float64)
	for _, backtag := range admin_weight {
		type_nums :=utils.GetInt64(strings.Split(backtag, ":")[0])
		admin_weight_num :=utils.GetFloat64(strings.Split(backtag, ":")[0])
		adminMap[type_nums] = admin_weight_num
	}

	//曝光逻辑
	for index := 0; index < ctx.GetDataLength(); index++ {
		randomScore := float32(rand.Intn(100)) / 100.0
		dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
		sd := dataInfo.SearchData//SearchData缺少类型信息,
		//rankInfo := dataInfo.GetRankInfo()

		itemScore:=randomScore*float32(sd.Weight)
		log.Infof("itemScore===============%+v", itemScore)
		//rankInfo.AddRecommend("sortScoreItem", itemScore)
	}
	return nil
}


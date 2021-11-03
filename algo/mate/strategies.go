package mate

import (
	"math/rand"
	"rela_recommend/algo"
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

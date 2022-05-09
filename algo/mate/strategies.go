package mate

import (
	"math/rand"
	"rela_recommend/algo"
	"rela_recommend/log"
)

// 内容较短，包含关键词的内容沉底
func BaseScoreStrategyItem(ctx algo.IContext, iDataInfo algo.IDataInfo, rankInfo *algo.RankInfo) error {
	abtest := ctx.GetAbTest()
	dataInfo := iDataInfo.(*DataInfo)
	sd := dataInfo.SearchData

	log.Infof("search===============%+v",sd)
	abSwitch := abtest.GetBool("mate_text_switch", false)
	if abSwitch {
		randomScore := float32(rand.Intn(100)) / 100.0
		rankInfo.Score = float32(sd.Weight) + randomScore
	}

	return nil
}
//
func SortScoreItem(ctx algo.IContext) error {
	var itemWeightMap = make(map[int64]int)
	for index := 0; index < ctx.GetDataLength(); index++ {
		dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
		sd := dataInfo.SearchData
		log.Infof("Map===============%+v",sd)
		itemWeightMap[sd.Id]=sd.Weight
	}
	log.Infof("searchMap===============%+v",itemWeightMap)
	return nil
}


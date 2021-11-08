package ad

import (
	"math"
	"math/rand"
	"rela_recommend/algo"
	"rela_recommend/log"
	rutils "rela_recommend/utils"
	"time"
)

// 内容较短，包含关键词的内容沉底
func BaseScoreStrategyItem(ctx algo.IContext, iDataInfo algo.IDataInfo, rankInfo *algo.RankInfo) error {
	abtest := ctx.GetAbTest()
	dataInfo := iDataInfo.(*DataInfo)
	sd := dataInfo.SearchData

	var priceRate float64 = 1.0
	if sd.Cpm > 0 { // 以默认 10元为基准1，100元为2，2.16为0.5
		priceRate = math.Log(float64(sd.Cpm) + 1.0)
	}

	var cntRate float64 = 1.0
	if sd.Exposure > 0 {
		runningRate := float64(ctx.GetCreateTime().Unix()-sd.StartTime) / float64(sd.EndTime-sd.StartTime)
		exposureRate := float64(sd.HistoryExposures) / float64(sd.Exposure)
		cnt_z := abtest.GetFloat64("base_score_cnt_z", 20.0)
		cntRate = math.Min(math.Pow(runningRate/exposureRate, cnt_z), 10000) // 无穷大
	}

	var clickRate float64 = 1.0
	var weightRate float64 = 1.0 + float64(sd.Weight)/100.0

	rankInfo.Score = float32(priceRate * cntRate * clickRate * weightRate)
	return nil
}
//广告分发策略
func BaseFeedPrice(ctx algo.IContext,iDataInfo algo.IDataInfo, rankInfo *algo.RankInfo) error{
	request := ctx.GetRequest()
	dataInfo := iDataInfo.(*DataInfo)
	sd := dataInfo.SearchData
	rand.Seed(time.Now().Unix())
	rand_num := rand.Int63n(5)+1.0
	nums:=float32(rand_num)/float32(sd.Id)
	log.Infof("rand_num===========",nums)
	if sd.Status == 1 && rutils.NewSetInt64FromArray(sd.TestUsers).Contains(request.UserId) {
		rankInfo.AddRecommend("ad_sort",1.0+float32(nums))
	}
	return nil
}


// 测试用户查看测试内容时置顶
func TestUserTopStrategyItem(ctx algo.IContext, iDataInfo algo.IDataInfo, rankInfo *algo.RankInfo) error {
	request := ctx.GetRequest()
	dataInfo := iDataInfo.(*DataInfo)
	sd := dataInfo.SearchData
	if sd.Status == 1 && rutils.NewSetInt64FromArray(sd.TestUsers).Contains(request.UserId) {
		rankInfo.IsTop = 1
	}
	return nil
}

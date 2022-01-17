package ad

import (
	"math"
	"math/rand"
	"rela_recommend/algo"
	"rela_recommend/log"
	rutils "rela_recommend/utils"
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
func BaseFeedPrice(ctx algo.IContext) error {
	request := ctx.GetRequest()
	dataLen:=ctx.GetDataLength()
	//召回的广告数据大于才做分发
	if request.ClientVersion>= 50802 && dataLen>1{
		app := ctx.GetAppInfo()
		dataLen:=ctx.GetDataLength()
		if ctx.GetDataLength() != 0 {
			for index := 0; index < ctx.GetDataLength(); index++ {
				dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
				rankInfo := dataInfo.GetRankInfo()
				sd := dataInfo.SearchData
				//		rand_num := rand.Intn(5) + 1.0
				//		nums := float32(rand_num) / float32(sd.Id)
				if app.Name=="ad.feed" {
						hisexpores := dataInfo.SearchData.HistoryExposures
						click := dataInfo.SearchData.HistoryClicks
						//rand_num := -(rand.Intn(5) + hisexpores)/dataLen
						if click > hisexpores {
							click = hisexpores
						}
						rand_num := rand.Intn(dataLen)
						ctr := float64(click+1) / float64(rand.Intn(dataLen)+hisexpores+1)
						nums := float64(ctr) * math.Exp(-float64(rand_num))
						log.Infof("sdId===============", sd.Id)
						log.Infof("click===============", click)
						log.Infof("hisexpores===============", hisexpores)
						log.Infof("rand_nums===============", ctr, nums)
						rankInfo.AddRecommend("ad_sort.feed", 1.0+float32(nums))
					}
				if app.Name=="ad.init"{
						hisexpores :=dataInfo.SearchData.HistoryExposures
						click :=dataInfo.SearchData.HistoryClicks
						if click>hisexpores{
							click=hisexpores
						}
						//rand_num := -(rand.Intn(5) + hisexpores)/dataLen
						rand_num := rand.Intn(dataLen)
						ctr:=float64(click+1)/float64(rand.Intn(dataLen) + hisexpores + 1)
						nums :=float64(ctr) * math.Exp(-float64(rand_num))
						log.Infof("sdId===============", sd.Id)
						log.Infof("click===============", click)
						log.Infof("hisexpores===============",hisexpores)
						log.Infof("rand_nums===============",rand_num,nums)
						rankInfo.AddRecommend("ad_sort.init", 1.0+float32(nums))
					}

				}
			}
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

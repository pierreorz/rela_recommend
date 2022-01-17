package ad

import (
	"math"
	"math/rand"
	"rela_recommend/algo"
	"rela_recommend/log"
	"rela_recommend/models/behavior"
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
func BaseFeedPrice(ctx algo.IContext,iDataInfo algo.IDataInfo, rankInfo *algo.RankInfo) error {
	request := ctx.GetRequest()
	dataLen:=ctx.GetDataLength()
	//召回的广告数据大于才做分发
	if request.ClientVersion>= 50802 && dataLen>1{
		params := ctx.GetRequest()
		behaviorCache := behavior.NewBehaviorCacheModule(ctx)
		app := ctx.GetAppInfo()
		var userBehavior *behavior.UserBehavior // 用户实时行为
		var userFeedId int64
		var userInitId int64
		dataLen:=ctx.GetDataLength()
		log.Infof("dataLen=================search_result_nums",dataLen)
		realtimes, realtimeErr := behaviorCache.QueryAdBehaviorMap("ad", []int64{params.UserId})
		log.Infof("realtimes=========================== %+v",realtimes)
		if realtimeErr == nil { // 获取flink数据
			userBehavior = realtimes[params.UserId]
			log.Infof("userBehavior=========================== %+v",userBehavior)
			if userBehavior != nil { //开屏广告和feed流广告id
				userFeedList := userBehavior.GetAdFeedListExposure().GetLastAdIds()
				userInitList := userBehavior.GetAdInitListExposure().GetLastAdIds()
				log.Infof("userFeedList=========================== %+v", userFeedList)
				log.Infof("userInitList=========================== %+v", userInitList)
				if len(userFeedList) > 0 {
					userFeedId = userFeedList[len(userFeedList)-1]
				}
				if len(userInitList) > 0 {
					userInitId = userInitList[len(userInitList)-1]
				}
				log.Infof("userFeedId=================userInitId",userFeedId,userInitId)
			}

		}
		dataInfo := iDataInfo.(*DataInfo)
		sd := dataInfo.SearchData
		//		rand_num := rand.Intn(5) + 1.0
		//		nums := float32(rand_num) / float32(sd.Id)
		if app.Name=="ad.feed" {
			if sd.Id != userFeedId {
				log.Infof("userFeedId===============",userFeedId)
				log.Infof("addWeigth_feed===============",sd.Id)
				hisexpores :=dataInfo.SearchData.HistoryExposures
				click :=dataInfo.SearchData.HistoryClicks
				//rand_num := -(rand.Intn(5) + hisexpores)/dataLen
				if click>hisexpores{
					click=hisexpores
				}
				rand_num := rand.Intn(dataLen)
				ctr:=float64(click+1)/float64(rand.Intn(dataLen) + hisexpores + 1)
				nums :=float64(ctr) * math.Exp(-float64(rand_num))
				log.Infof("hisexpores===============",hisexpores)
				log.Infof("rand_nums===============",ctr,nums)
				rankInfo.AddRecommend("ad_sort.feed", 1.0+float32(nums))
			}else{
				nums := math.Exp(-float64(dataLen))
				rankInfo.AddRecommend("ad_sort.feed", 1.0-float32(nums))
			}
		}
		if app.Name=="ad.init"{
			if sd.Id != userInitId {
				log.Infof("userInitId===============",userInitId)
				log.Infof("addWeigth_init===============",sd.Id)
				hisexpores :=dataInfo.SearchData.HistoryExposures
				click :=dataInfo.SearchData.HistoryClicks
				if click>hisexpores{
					click=hisexpores
				}
				//rand_num := -(rand.Intn(5) + hisexpores)/dataLen
				rand_num := rand.Intn(dataLen)
				ctr:=float64(click+1)/float64(rand.Intn(dataLen) + hisexpores + 1)
				nums :=float64(ctr) * math.Exp(-float64(rand_num))
				log.Infof("hisexpores===============",hisexpores)
				log.Infof("rand_nums===============",rand_num,nums)
				rankInfo.AddRecommend("ad_sort.init", 1.0+float32(nums))
			}else{
				nums := math.Exp(-float64(dataLen))
				rankInfo.AddRecommend("ad_sort.feed", 1.0-float32(nums))
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

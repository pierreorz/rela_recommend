package live

import (
	"math"
	"math/rand"
	"rela_recommend/algo"
)

// 处理业务给出的置顶和推荐内容
func LiveTopRecommandStrategyFunc(ctx algo.IContext, index int) error {
	dataInfo := ctx.GetDataByIndex(index)
	rankInfo := dataInfo.GetRankInfo()
	live := dataInfo.(*LiveInfo)

	if live.LiveCache.Recommand == 1 { // 1: 推荐
		if live.LiveCache.RecommandLevel > 10 { // 15: 置顶
			rankInfo.IsTop = 1
		}
		rankInfo.Level = live.LiveCache.RecommandLevel // 推荐等级
	} else if live.LiveCache.Recommand == -1 { // -1: 不推荐
		if live.LiveCache.RecommandLevel == -1 { // -1: 置底
			rankInfo.IsTop = -1
		} else if live.LiveCache.RecommandLevel > 0 { // 降低权重
			level := math.Min(float64(live.LiveCache.RecommandLevel), 100.0)
			// rankInfo.Punish = float32(100.0 - level) / 100.0
			rankInfo.AddRecommend("Down", float32(100.0-level)/100.0)
		}
	}
	return nil
}

// 融合老策略的分数
type OldScoreStrategy struct{}

func (self *OldScoreStrategy) Do(ctx algo.IContext) error {
	var err error
	new_score := ctx.GetAbTest().GetFloat("new_score", 1.0)
	old_score := 1 - new_score
	for i := 0; i < ctx.GetDataLength(); i++ {
		dataInfo := ctx.GetDataByIndex(i)
		live := dataInfo.(*LiveInfo)
		rankInfo := dataInfo.GetRankInfo()
		score := self.oldScore(live)
		rankInfo.Score = live.RankInfo.Score*new_score + score*old_score
	}
	return err
}
func (self *OldScoreStrategy) scoreFx(score float32) float32 {
	return (score / 200) / (1 + score/200)
}
func (self *OldScoreStrategy) oldScore(live *LiveInfo) float32 {
	var score float32 = 0
	score += self.scoreFx(live.LiveCache.DayIncoming) * 0.2
	score += self.scoreFx(live.LiveCache.MonthIncoming) * 0.05
	score += self.scoreFx(live.LiveCache.Score) * 0.55
	score += self.scoreFx(float32(live.LiveCache.FansCount)) * 0.10
	score += self.scoreFx(float32(live.LiveCache.Live.ShareCount)) * 0.10
	return score
}

// 对于上个小时榜前3名进行随机制前
func HourRankRecommendFunc(ctx algo.IContext) error {
	abtest := ctx.GetAbTest()
	topN := abtest.GetInt("per_hour_rank_top_n", 3) // 前n名随机， 分数相同的并列，有可能返回1,2,2,3
	indexs := []int{}
	for index := 0; index < ctx.GetDataLength(); index++ {
		dataInfo := ctx.GetDataByIndex(index).(*LiveInfo)
		if dataInfo.LiveData != nil && dataInfo.LiveData.PreHourRank > 0 && dataInfo.LiveData.PreHourRank <= topN {
			indexs = append(indexs, index)
		}
	}
	if len(indexs) > 0 {
		index := rand.Intn(len(indexs))
		liveInfo := ctx.GetDataByIndex(index).(*LiveInfo)
		rankInfo := liveInfo.GetRankInfo()
		rankInfo.IsTop = 1
		rankInfo.Level = -1
	}
	return nil
}

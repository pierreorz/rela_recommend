package live

import(
	"math"
	"rela_recommend/algo"
)

// 按照6小时优先策略
func LiveTopRecommandStrategyFunc(ctx algo.IContext, index int) error {
	dataInfo := ctx.GetDataByIndex(index)
	rankInfo := dataInfo.GetRankInfo()
	live := dataInfo.(*LiveInfo)

	if live.LiveCache.Recommand == 1 {							// 1: 推荐
		if live.LiveCache.RecommandLevel > 10 {						// 15: 置顶
			rankInfo.IsTop = 1
		}
		rankInfo.Level = live.LiveCache.RecommandLevel			// 推荐等级
	}  else if live.LiveCache.Recommand == -1 {					// -1: 不推荐
		if live.LiveCache.RecommandLevel == -1 {					// -1: 置底
			rankInfo.IsTop = -1
		} else if live.LiveCache.RecommandLevel > 0 {				// 降低权重
			level := math.Min(float64(live.LiveCache.RecommandLevel), 100.0)
			rankInfo.Punish = float32(100.0 - level) / 100.0
		}
	}
	return nil
}

// 融合老策略的分数
type OldScoreStrategy struct { }
func (self *OldScoreStrategy) Do(ctx algo.IContext) error {
	var err error
	new_score := ctx.GetAbTest().GetFloat("new_score", 1.0)
	old_score := 1 - new_score
	for i := 0; i < ctx.GetDataLength(); i++ {
		dataInfo := ctx.GetDataByIndex(i)
		live := dataInfo.(*LiveInfo)
		rankInfo := dataInfo.GetRankInfo()
		score := self.oldScore(live)
		rankInfo.Score = live.RankInfo.Score * new_score + score * old_score
	}
	return err
}
func (self *OldScoreStrategy) scoreFx(score float32) float32 {
	return (score / 200) / (1 + score / 200)
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

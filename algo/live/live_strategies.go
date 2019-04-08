package live

import (
	"math"
)

type LiveStrategy interface {
	Do(ctx *LiveAlgoContext, list []LiveInfo)
}

type LiveStrategyBase struct {}

// 业务置顶推荐惩罚策略
type LiveTopStrategy struct { LiveStrategyBase } 
func (self *LiveTopStrategy) Do(ctx *LiveAlgoContext, list []LiveInfo) {
	for i, _ := range list { 
		live := &list[i]
		if live.LiveCache.Recommand == 1 {							// 1: 推荐
			if live.LiveCache.RecommandLevel > 10 {						// 15: 置顶
				live.RankInfo.IsTop = 1
			}
			live.RankInfo.Level = live.LiveCache.RecommandLevel			// 推荐等级
		}  else if live.LiveCache.Recommand == -1 {					// -1: 不推荐
			if live.LiveCache.RecommandLevel == -1 {					// -1: 置底
				live.RankInfo.IsTop = -1
			} else if live.LiveCache.RecommandLevel > 0 {				// 降低权重
				level := math.Min(float64(live.LiveCache.RecommandLevel), 100.0)
				live.RankInfo.Punish = float32(100.0 - level) / 100.0
			}
		}
	}
}

// 降权策略
type LivePunishStrategy struct { LiveStrategyBase } 
func (self *LivePunishStrategy) Do(ctx *LiveAlgoContext, list []LiveInfo) {
	for i, _ := range list { 
		live := &list[i]
		if live.RankInfo.Punish > 0.000001 {
			live.RankInfo.Score = live.RankInfo.Score * live.RankInfo.Punish
		}
	}
}

// 融合老策略
type LiveOldStrategy struct { LiveStrategyBase }
func (self *LiveOldStrategy) Do(ctx *LiveAlgoContext, list []LiveInfo) {
	new_score := ctx.AbTest.GetFloat("new_score", 1.0)
	old_score := 1 - new_score
	for i, _ := range list {
		live := &list[i]
		score := self.oldScore(ctx, live) 
		live.RankInfo.Score = live.RankInfo.Score * new_score + score * old_score
	}
}
func (self *LiveOldStrategy) scoreFx(score float32) float32 {
	return (score / 200) / (1 + score / 200)
}
func (self *LiveOldStrategy) oldScore(ctx *LiveAlgoContext, live *LiveInfo) float32 {
	var score float32 = 0
	score += self.scoreFx(live.LiveCache.DayIncoming) * 0.2
	score += self.scoreFx(live.LiveCache.MonthIncoming) * 0.05
	score += self.scoreFx(live.LiveCache.Score) * 0.55
	score += self.scoreFx(float32(live.LiveCache.FansCount)) * 0.10
	score += self.scoreFx(float32(live.LiveCache.Live.ShareCount)) * 0.10
	return score
}

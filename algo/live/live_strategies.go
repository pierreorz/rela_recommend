package live

type LiveStrategy interface {
	Do(ctx *LiveAlgoContext, list []LiveInfo)
	// DoSingle(ctx *LiveAlgoContext, live *LiveInfo)
}

type LiveStrategyBase struct {}

// 置顶策略
type LiveTopStrategy struct { LiveStrategyBase } 
func (self *LiveTopStrategy) Do(ctx *LiveAlgoContext, list []LiveInfo) {
	for i, _ := range list { self.DoSingle(ctx, &list[i]) }
}
func (self *LiveTopStrategy) DoSingle(ctx *LiveAlgoContext, live *LiveInfo) {
	// 1: 置顶， 0: 默认， -1:置底
	if live.LiveCache.Recommand == 1 && live.LiveCache.RecommandLevel > 10 {
		live.RankInfo.IsTop = 1
	} else if live.LiveCache.Recommand == -1 && live.LiveCache.RecommandLevel == -1 {
		live.RankInfo.IsTop = -1
	} else {
		live.RankInfo.IsTop = 0
	}
}

// 等级策略
type LiveLevelStrategy struct { LiveStrategyBase } 
func (self *LiveLevelStrategy) Do(ctx *LiveAlgoContext, list []LiveInfo) {
	for i, _ := range list { self.DoSingle(ctx, &list[i]) }
}
func (self *LiveLevelStrategy) DoSingle(ctx *LiveAlgoContext, live *LiveInfo) {
	live.RankInfo.Level = live.LiveCache.RecommandLevel
}

// 融合老策略
type LiveOldStrategy struct { LiveStrategyBase }
func (self *LiveOldStrategy) Do(ctx *LiveAlgoContext, list []LiveInfo) {
	for i, _ := range list { self.DoSingle(ctx, &list[i]) }
}
func (self *LiveOldStrategy) scoreFx(score float32) float32 {
	return (score / 200) / (1 + score / 200)
}
func (self *LiveOldStrategy) DoSingle(ctx *LiveAlgoContext, live *LiveInfo) {
	var score float32 = 0
	score += self.scoreFx(live.LiveCache.DayIncoming) * 0.2
	score += self.scoreFx(live.LiveCache.MonthIncoming) * 0.05
	score += self.scoreFx(live.LiveCache.Score) * 0.55
	score += self.scoreFx(float32(live.LiveCache.FansCount)) * 0.10
	score += self.scoreFx(float32(live.LiveCache.Live.ShareCount)) * 0.10

	live.RankInfo.Score = live.RankInfo.Score * 0.7 + score * 0.3
}

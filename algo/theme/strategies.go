package theme

import(
	"math"
	"rela_recommend/algo"
)

// 热门提升权重
func DoHotBehaviorUpper(ctx algo.IContext, index int) error {
	dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
	rankInfo := dataInfo.GetRankInfo()
	var avgCount float32 = 1000
	var upperRate float32 = 0.2
	behavior := dataInfo.ThemeBehavior
	if behavior != nil {
		if behavior.IsListExposured() {
			countRate := 2.0 / (1 + math.Exp(-float64(behavior.ListExposure.Count / avgCount))) -1
			upperRate = behavior.ListClickRate() * float32(countRate)
		}
	}
	rankInfo.Score = rankInfo.Score * (1.0 + upperRate)
	return nil
}

// 对自己的行为进行权重处理
func DoUserBehaviorUpper(ctx algo.IContext, index int) error {
	dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
	rankInfo := dataInfo.GetRankInfo()

	var avgCount float32 = 10
	var upperRate float32 = 1.0
	behavior := dataInfo.ThemeBehavior
	if behavior != nil {
		if behavior.IsListExposured() {
			countRate := 2.0 / (1 + math.Exp(-float64(behavior.ListExposure.Count / avgCount))) -1
			clickRate := behavior.ListClickRate()
			if clickRate <= 0.000001 {	// 没有点击直接降权
				upperRate = -float32(countRate)
			} else {
				upperRate = behavior.ListClickRate() * float32(countRate)
			}
		}
	}
	rankInfo.Score = rankInfo.Score * (1.0 + upperRate)
	return nil
}
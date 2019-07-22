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
			count2 := float64(behavior.ListExposure.Count / avgCount)
			countRate := count2 / (1 + count2)
			upperRate = 10 * behavior.ListClickRate() * float32(countRate)
		}
	}
	rankInfo.Score = rankInfo.Score * (1.0 + upperRate)
	return nil
}

// 对自己的行为进行权重处理
type UserBehaviorStrategy struct { }
func (self *UserBehaviorStrategy) Do(ctx algo.IContext) error {
	var err error
	var avgCount float32 = 1
	var upperRate float32 = 0.2
	var currTime = float32(ctx.GetCreateTime().Unix())
	for index := 0; index < ctx.GetDataLength(); index++ {
		dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
		rankInfo := dataInfo.GetRankInfo()

		behavior := dataInfo.ThemeBehavior
		if behavior != nil {
			if behavior.ListExposure != nil && behavior.ListExposure.Count > 0 {
				// 展示次数
				count2 := float64(behavior.ListExposure.Count / avgCount)
				countRate := count2 / (1 + count2)
				clickRate := behavior.ListClickRate()
				if clickRate <= 0.000001 {	// 没有点击直接降权
					timeSpc := 1 / (1 + math.Abs(float64(currTime - behavior.ListExposure.LastTime)) / 60.0)
					upperRate = -2 * float32(countRate) * float32(timeSpc)
				} else {
					timeSpc := 1 / (1 + math.Abs(float64(currTime - behavior.ListClick.LastTime)) / 600.0)
					upperRate = clickRate * float32(countRate) * float32(timeSpc)
				}
			}
		}
		rankInfo.Score = rankInfo.Score * (1.0 + upperRate)
	}
	return err
}
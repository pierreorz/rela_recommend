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
type UserBehaviorStrategy struct { }
func (self *UserBehaviorStrategy) Do(ctx algo.IContext) error {
	var err error
	var avgCount float32 = 5
	var upperRate float32 = 0.5
	var currTime = float32(ctx.GetCreateTime().Unix())
	for index := 0; index < ctx.GetDataLength(); index++ {
		dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
		rankInfo := dataInfo.GetRankInfo()

		behavior := dataInfo.ThemeBehavior
		if behavior != nil {
			if behavior.IsListExposured() {
				// 展示次数
				countRate := 2.0 / (1 + math.Exp(-float64(behavior.ListExposure.Count / avgCount))) -1
				clickRate := behavior.ListClickRate()
				if clickRate <= 0.000001 {	// 没有点击直接降权
					timeSpc := 1 / (1 + math.Abs(float64(currTime - behavior.ListExposure.LastTime)) / 300.0)
					upperRate = -float32(countRate) * float32(timeSpc)
				} else {
					timeSpc := 1 / (1 + math.Abs(float64(currTime - behavior.ListClick.LastTime)) / 300.0)
					upperRate = behavior.ListClickRate() * float32(countRate) * float32(timeSpc)
				}
			}
		}
		rankInfo.Score = rankInfo.Score * (1.0 + upperRate)
	}
	return err
}
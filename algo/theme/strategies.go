package theme

import(
	"math"
	"rela_recommend/algo"
	"rela_recommend/algo/utils"
	rutils "rela_recommend/utils"
)

// 热门提升权重
func DoHotBehaviorUpper(ctx algo.IContext, index int) error {
	dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
	rankInfo := dataInfo.GetRankInfo()
	var avgCount float64 = 1000
	var upperRate float32 = 0.1
	behavior := dataInfo.ThemeBehavior
	if behavior != nil {
		if behavior.ListExposure != nil && behavior.ListExposure.Count > 0 {
			clickRate := math.Max(0.000001, math.Min(1.0, behavior.ListClickRate()))

			countScore := 1.0 - math.Exp(- behavior.ListExposure.Count / avgCount)
			clickScore := utils.ExpLogit(clickRate)
			upperRate = float32(clickScore * countScore)
		}
	}
	rankInfo.AddRecommend("ThemeBehavior", 1.0 + upperRate)
	
	return nil
}

// 对自己的行为进行权重处理
type UserBehaviorStrategy struct { }
func (self *UserBehaviorStrategy) Do(ctx algo.IContext) error {
	var err error
	var avgCount float64 = 2
	var upperRate float32 = 0.1
	var currTime = float64(ctx.GetCreateTime().Unix())
	for index := 0; index < ctx.GetDataLength(); index++ {
		dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
		rankInfo := dataInfo.GetRankInfo()

		behavior := dataInfo.UserBehavior
		if behavior != nil {
			if behavior.ListExposure != nil && behavior.ListExposure.Count > 0 {
				clickRate := math.Max(0.000001, math.Min(1.0, behavior.ListClickRate()))
				lastTime := math.Max(behavior.ListExposure.LastTime, behavior.ListClick.LastTime)
				lastIsClick := behavior.ListClick.LastTime >= behavior.ListExposure.LastTime
				
				var lastNotClick float64 = rutils.IfElse(lastIsClick, 0.0, 1.0)		// 最后一次是否点击
				var timeBase float64 = rutils.IfElse(lastIsClick, 600.0, 60.0)		// 时间衰减速度

				countScore := 1.0 - math.Exp(- behavior.ListExposure.Count / avgCount)
				timeScore := math.Exp(- (currTime - lastTime) / timeBase)
				clickScore := 2 * utils.ExpLogit(clickRate) - lastNotClick
				upperRate =  float32(clickScore * countScore * timeScore)
			}
		}
		rankInfo.AddRecommend("UserBehavior", 1.0 + upperRate)
	}
	return err
}

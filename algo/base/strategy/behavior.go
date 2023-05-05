package strategy

import (
	"rela_recommend/algo"
	"rela_recommend/models/behavior"
)

// 数据行为处理策略
type BaseBehaviorRichStrategy struct {
	ctx         algo.IContext
	cacheModule *behavior.BehaviorCacheModule

	DefaultWeight   int
	UserBehaviorMap map[int64]*behavior.UserBehavior
	ItemBehaviorMap map[int64]*behavior.UserBehavior

	UserStrategyFunc     func(algo.IContext, map[int64]*behavior.UserBehavior) error
	UserStrategyItemFunc func(algo.IContext, algo.IDataInfo, *behavior.UserBehavior, *algo.RankInfo) error

	ItemStrategyFunc     func(algo.IContext, map[int64]*behavior.UserBehavior) error
	ItemStrategyItemFunc func(algo.IContext, algo.IDataInfo, *behavior.UserBehavior, *algo.RankInfo) error
}

func (self *BaseBehaviorRichStrategy) GetDefaultWeight() int {
	return self.DefaultWeight
}

func (self *BaseBehaviorRichStrategy) New(ctx algo.IContext) algo.IRichStrategy {
	return &BaseBehaviorRichStrategy{
		ctx:                  ctx,
		cacheModule:          behavior.NewBehaviorCacheModule(ctx),
		UserBehaviorMap:      map[int64]*behavior.UserBehavior{},
		ItemBehaviorMap:      map[int64]*behavior.UserBehavior{},
		UserStrategyFunc:     self.UserStrategyFunc,
		UserStrategyItemFunc: self.UserStrategyItemFunc,
		ItemStrategyFunc:     self.ItemStrategyFunc,
		ItemStrategyItemFunc: self.ItemStrategyItemFunc}
}
func (self *BaseBehaviorRichStrategy) BuildData() error {
	app := self.ctx.GetAppInfo()
	params := self.ctx.GetRequest()
	if userBehavior, err := self.cacheModule.QueryUserItemBehaviorMap(
		app.Module, params.UserId, self.ctx.GetDataIds()); err == nil {
		self.UserBehaviorMap = userBehavior
	}
	if itemBehavior, err := self.cacheModule.QueryItemBehaviorMap(
		app.Module, self.ctx.GetDataIds()); err == nil {
		self.ItemBehaviorMap = itemBehavior
	}
	return nil
}

func (self *BaseBehaviorRichStrategy) Strategy() error {
	var err error
	if self.UserStrategyFunc != nil && self.UserBehaviorMap != nil {
		err = self.UserStrategyFunc(self.ctx, self.UserBehaviorMap)
	}
	if self.ItemStrategyFunc != nil && self.ItemBehaviorMap != nil {
		err = self.ItemStrategyFunc(self.ctx, self.UserBehaviorMap)
	}
	if self.UserStrategyItemFunc != nil || self.ItemStrategyItemFunc != nil {
		for index := 0; index < self.ctx.GetDataLength(); index++ {
			dataInfo := self.ctx.GetDataByIndex(index)
			dataId := dataInfo.GetDataId()
			rankInfo := dataInfo.GetRankInfo()
			if self.UserBehaviorMap != nil {
				behavior, _ := self.UserBehaviorMap[dataId]
				self.UserStrategyItemFunc(self.ctx, dataInfo, behavior, rankInfo)
			}
			if self.ItemBehaviorMap != nil {
				if behavior, ok := self.ItemBehaviorMap[dataId]; ok && behavior != nil {
					self.ItemStrategyItemFunc(self.ctx, dataInfo, behavior, rankInfo)
				}
			}
		}
	}
	return err
}

func (self *BaseBehaviorRichStrategy) Logger() error {
	return nil
}

// 对于曝光不足的内容进行加权曝光
func ExposureIncreaseFunc(ctx algo.IContext) error {
	abtest := ctx.GetAbTest()
	increaseThreshold := abtest.GetFloat64("exposure_increase_threshold", 0.0) // 需要提升的曝光阈值，曝光小于该值才会增加曝光
	increaseMax := abtest.GetFloat64("exposure_increase_max", 0.2)             // 最多增加的分数
	increaseExposures := abtest.GetStrings("exposure_increase_exposures", "around.list:exposure")
	if increaseThreshold > 0.0 && increaseMax > 0.0 && len(increaseExposures) > 0 {
		for index := 0; index < ctx.GetDataLength(); index++ {
			dataInfo := ctx.GetDataByIndex(index)
			rankInfo := dataInfo.GetRankInfo()

			if itemBehavior := dataInfo.GetBehavior(); itemBehavior != nil {
				exposures := itemBehavior.Gets(increaseExposures...)
				if exposures.Count < increaseThreshold { // 曝光不足提权
					score := float32((increaseThreshold - exposures.Count) / increaseThreshold * increaseMax)
					rankInfo.AddRecommend("ExposureIncrease", 1+score)
				}
			}
		}
	}
	return nil
}

// 给足曝光但无互动的降权
func NoInteractDecreaseFunc(ctx algo.IContext) error {
	abtest := ctx.GetAbTest()
	decreaseScore := abtest.GetFloat("no_interact_decrease_max", 0.2)                    // 最多减小的分数
	interactBehaviors := abtest.GetStrings("interact_behaviors", "around.list:click")    // 互动行为
	exposureBehaviors := abtest.GetStrings("exposure_behaviors", "around.list:exposure") // 曝光行为
	exposureThreshold := abtest.GetFloat64("success_exposure_threshold", 0.0)            // 足额曝光的阈值
	if decreaseScore > 0.0 && len(interactBehaviors) > 0 && len(exposureBehaviors) > 0 && exposureThreshold > 0 {
		for index := 0; index < ctx.GetDataLength(); index++ {
			dataInfo := ctx.GetDataByIndex(index)
			rankInfo := dataInfo.GetRankInfo()

			if itemBehavior := dataInfo.GetBehavior(); itemBehavior != nil {
				actions := itemBehavior.Gets(interactBehaviors...)
				exposures := itemBehavior.Gets(exposureBehaviors...)
				if exposures.Count >= exposureThreshold && actions.Count <= 0 {
					rankInfo.AddRecommend("NoInteractDecrease", 1-decreaseScore)
				}
			}
		}
	}
	return nil
}

// 曝光后沉底
func ExposureBottomFunc(ctx algo.IContext) error {
	abtest := ctx.GetAbTest()
	interactBehaviors := abtest.GetStrings("interact_behaviors", "around.list:click")    // 互动行为
	exposureBehaviors := abtest.GetStrings("exposure_behaviors", "around.list:exposure") // 曝光行为
	if len(interactBehaviors) > 0 && len(exposureBehaviors) > 0 {
		for index := 0; index < ctx.GetDataLength(); index++ {
			dataInfo := ctx.GetDataByIndex(index)
			rankInfo := dataInfo.GetRankInfo()

			if userItemBehavior := dataInfo.GetUserBehavior(); userItemBehavior != nil {
				actions := userItemBehavior.Gets(interactBehaviors...)
				exposures := userItemBehavior.Gets(exposureBehaviors...)
				if exposures.Count > 0 || actions.Count > 0 {
					rankInfo.AddRecommend("ExposureBottom", 0.01)
				}
			}
		}
	}
	return nil
}



func FlowPacketFunc(ctx algo.IContext) error{
	abtest := ctx.GetAbTest()
	exposureBehaviors :=abtest.GetStrings("exposure_behaviors","moment.recommend:exposure")
	for index :=0 ; index <ctx.GetDataLength(); index ++{
		dataInfo := ctx.GetDataByIndex(index)
		rankInfo := dataInfo.GetRankInfo()
		if rankInfo.Packet>0 && rankInfo.IsTarget>0{
			count :=0.0
			if userItemBehavior := dataInfo.GetUserBehavior(); userItemBehavior != nil {
				exposures := userItemBehavior.Gets(exposureBehaviors...)
				count = exposures.Count

			}else{
				count = 0.0
			}
			rankInfo.AddRecommend("recommend_plan",float32(1+0.2*(1-count/rankInfo.Packet)))
		}
	}
	return nil
}

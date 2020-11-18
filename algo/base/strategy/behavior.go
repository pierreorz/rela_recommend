package strategy

import (
	"rela_recommend/algo"
	"rela_recommend/factory"
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
		cacheModule:          behavior.NewBehaviorCacheModule(ctx, &factory.CacheBehaviorRds),
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

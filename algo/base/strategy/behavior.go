package strategy


import (
	"rela_recommend/log"
	"rela_recommend/algo"
	"rela_recommend/factory"
	"rela_recommend/models/behavior"
)


// 数据行为处理策略
type BaseBehaviorRichStrategy struct {
	ctx					algo.IContext
	cacheModule				*behavior.BehaviorCacheModule
	UserBehaviorMap			map[int64]*behavior.UserBehavior
	ItemBehaviorMap			map[int64]*behavior.UserBehavior

	UserStrategyFunc		func(algo.IContext, map[int64]*behavior.UserBehavior) error
	UserStrategyItemFunc	func(algo.IContext, *behavior.UserBehavior, *algo.RankInfo) error

	ItemStrategyFunc		func(algo.IContext, map[int64]*behavior.UserBehavior) error
	ItemStrategyItemFunc	func(algo.IContext, *behavior.UserBehavior, *algo.RankInfo) error
}

func (self *BaseBehaviorRichStrategy) New(ctx algo.IContext) algo.IRichStrategy {
	return &BaseBehaviorRichStrategy{
		ctx: ctx, 
		cacheModule: behavior.NewBehaviorCacheModule(ctx, &factory.CacheBehaviorRds),
		UserBehaviorMap: map[int64]*behavior.UserBehavior{},
		ItemBehaviorMap: map[int64]*behavior.UserBehavior{},
		UserStrategyFunc: self.UserStrategyFunc,
		UserStrategyItemFunc: self.UserStrategyItemFunc,
		ItemStrategyFunc: self.ItemStrategyFunc,
		ItemStrategyItemFunc: self.ItemStrategyItemFunc}
}
func (self *BaseBehaviorRichStrategy) BuildData() error {
	app := self.ctx.GetAppInfo()
	params := self.ctx.GetRequest()
	if userBehavior, err := self.cacheModule.QueryUserBehaviorMap(
			app.Module, params.UserId, self.ctx.GetDataIds()); err != nil {
		self.UserBehaviorMap =	userBehavior
	}
	if itemBehavior, err := self.cacheModule.QueryItemBehaviorMap(
			app.Module, self.ctx.GetDataIds()); err != nil {
		self.ItemBehaviorMap =	itemBehavior
	}
	return nil
}

func (self *BaseBehaviorRichStrategy) Strategy() error {
	log.Infof("BaseBehaviorRichStrategy start\n")
	if self.UserStrategyFunc != nil && self.UserBehaviorMap != nil {
		return self.UserStrategyFunc(self.ctx, self.UserBehaviorMap)
	}
	if self.ItemStrategyFunc != nil && self.ItemBehaviorMap != nil {
		return self.ItemStrategyFunc(self.ctx, self.UserBehaviorMap)
	}
	if self.UserStrategyItemFunc != nil || self.ItemStrategyItemFunc != nil {
		for index := 0; index < self.ctx.GetDataLength(); index++ {
			dataInfo := self.ctx.GetDataByIndex(index)
			dataId := dataInfo.GetDataId()
			rankInfo := dataInfo.GetRankInfo()
			if self.UserBehaviorMap != nil {
				if behavior, ok := self.UserBehaviorMap[dataId]; ok && behavior != nil {
					self.UserStrategyItemFunc(self.ctx, behavior, rankInfo)
					log.Infof("BaseBehaviorRichStrategy user item %+v %+v\n, rankInfo", behavior, rankInfo)
				}
			}
			if self.ItemBehaviorMap != nil {
				if behavior, ok := self.ItemBehaviorMap[dataId]; ok && behavior != nil {
					self.ItemStrategyItemFunc(self.ctx, behavior, rankInfo)
					log.Infof("BaseBehaviorRichStrategy user item %+v %+v\n, rankInfo", behavior, rankInfo)
				}
			}
		}
	}
	log.Infof("BaseBehaviorRichStrategy end\n")
	return nil
}

func (self *BaseBehaviorRichStrategy) Logger() error {
	return nil
}

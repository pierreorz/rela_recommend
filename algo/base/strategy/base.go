package strategy

import (
	"rela_recommend/algo"
)

// 数据行为处理策略
type BaseRichStrategy struct {
	ctx algo.IContext

	DefaultWeight 	int

	BuildDataFunc func(algo.IContext) error

	StrategyFunc     func(algo.IContext) error
	StrategyItemFunc func(algo.IContext, algo.IDataInfo, *algo.RankInfo) error

	LoggerFunc     func(algo.IContext) error
	LoggerItemFunc func(algo.IContext, algo.IDataInfo, *algo.RankInfo) error
}

func (self *BaseRichStrategy) New(ctx algo.IContext) algo.IRichStrategy {
	return &BaseRichStrategy{
		ctx:              ctx,
		BuildDataFunc:    self.BuildDataFunc,
		StrategyFunc:     self.StrategyFunc,
		StrategyItemFunc: self.StrategyItemFunc,
		LoggerFunc:       self.LoggerFunc,
		LoggerItemFunc:   self.LoggerItemFunc,
	}
}

func (self *BaseRichStrategy) GetDefaultWeight() int {
	return self.DefaultWeight
}

func (self *BaseRichStrategy) BuildData() error {
	if self.BuildDataFunc != nil {
		return self.BuildDataFunc(self.ctx)
	}
	return nil
}

func (self *BaseRichStrategy) Strategy() error {
	var err error
	if self.StrategyFunc != nil {
		err = self.StrategyFunc(self.ctx)
	}
	if self.StrategyItemFunc != nil {
		for index := 0; index < self.ctx.GetDataLength(); index++ {
			dataInfo := self.ctx.GetDataByIndex(index)
			rankInfo := dataInfo.GetRankInfo()
			self.StrategyItemFunc(self.ctx, dataInfo, rankInfo)
		}
	}
	return err
}

func (self *BaseRichStrategy) Logger() error {
	var err error
	if self.LoggerFunc != nil {
		err = self.LoggerFunc(self.ctx)
	}
	if self.LoggerItemFunc != nil {
		for index := 0; index < self.ctx.GetDataLength(); index++ {
			dataInfo := self.ctx.GetDataByIndex(index)
			rankInfo := dataInfo.GetRankInfo()
			self.LoggerItemFunc(self.ctx, dataInfo, rankInfo)
		}
	}
	return err
}

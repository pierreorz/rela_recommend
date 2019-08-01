package base

import (
	"fmt"
	"rela_recommend/algo"
	"rela_recommend/factory"
	"rela_recommend/models/redis"
)

type IRichStrategy interface {
	New(ctx algo.IContext) IRichStrategy
	BuildData() error		// 加载数据
	Strategy() error		// 执行策略
	Logger() error			// 记录结果
}


type PagedRichStrategy struct {
	ctx				algo.IContext
	pageIdsMap		map[int64]int
	cacheModule		*redis.ThemeBehaviorCacheModule
}

func (self *PagedRichStrategy) New(ctx algo.IContext) IRichStrategy {
	return &PagedRichStrategy{ctx: ctx}
}

func (self *PagedRichStrategy) CacheKey() string {
	abtest := self.ctx.GetAbTest()
	return fmt.Sprintf("algo:paged:%s:%d", abtest.App, abtest.DataId)
}

func (self *PagedRichStrategy) BuildData() error {
	self.cacheModule = redis.NewThemeBehaviorCacheModule(self.ctx, &factory.CacheBehaviorRds)
	params := self.ctx.GetRequest()
	if params.Offset > 0 {		// 只有非第一页才获取缓存
		return self.cacheModule.GetStruct(self.CacheKey(), self.pageIdsMap)
	}
	return nil
}

func (self *PagedRichStrategy) Strategy() error {
	for index := 0; index < self.ctx.GetDataLength(); index++ {
		dataInfo := self.ctx.GetDataByIndex(index)
		rankInfo := dataInfo.GetRankInfo()
		dataId := dataInfo.GetDataId()
		if value, ok := self.pageIdsMap[dataId]; ok {
			rankInfo.IsPaged = value
		}
	}
	return nil
}


func (self *PagedRichStrategy) Logger() error {
	if response := self.ctx.GetResponse(); response != nil {
		for _, item := range response.DataIds {
			self.pageIdsMap[item] = 1
		}
	}
	
	if len(self.pageIdsMap) > 0 {
		return self.cacheModule.SetStruct(self.CacheKey(), self.pageIdsMap, 30 * 60, 0)
	}
	return nil
}



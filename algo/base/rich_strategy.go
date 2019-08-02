package base

import (
	"fmt"
	"rela_recommend/log"
	"rela_recommend/algo"
	"rela_recommend/factory"
	"rela_recommend/models/redis"
)

// 分页历史处理策略
type PagedRichStrategy struct {
	ctx				algo.IContext
	pageIdsMap		map[int64]int
	cacheModule		*redis.CachePikaModule
}

func (self *PagedRichStrategy) New(ctx algo.IContext) algo.IRichStrategy {
	return &PagedRichStrategy{
		ctx: ctx, 
		pageIdsMap: map[int64]int{},
		cacheModule: redis.NewCachePikaModule(ctx, factory.CacheBehaviorRds)}
}

func (self *PagedRichStrategy) CacheKey() string {
	abtest := self.ctx.GetAbTest()
	return fmt.Sprintf("algo:paged:%s:%d", abtest.App, abtest.DataId)
}

func (self *PagedRichStrategy) BuildData() error {
	abtest := self.ctx.GetAbTest()
	params := self.ctx.GetRequest()
	if params.Offset > 0 {		// 只有非第一页才获取缓存
		self.cacheModule.GetStruct(self.CacheKey(), &self.pageIdsMap)
	}
	log.Infof("%s paged load %d len %d", abtest.App, params.Offset, len(self.pageIdsMap))
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
	abtest := self.ctx.GetAbTest()
	params := self.ctx.GetRequest()
	if response := self.ctx.GetResponse(); response != nil {
		for _, item := range response.DataIds {
			self.pageIdsMap[item] = 1
		}
	}
	
	if len(self.pageIdsMap) > 0 {
		go func(){
			self.cacheModule.SetStruct(self.CacheKey(), self.pageIdsMap, 30 * 60, 0)
		}()
	}
	log.Infof("%s paged save %d len %d", abtest.App, params.Offset, len(self.pageIdsMap))
	return nil
}

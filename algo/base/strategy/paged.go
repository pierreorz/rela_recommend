package strategy

import (
	"fmt"
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
	params := self.ctx.GetRequest()
	return fmt.Sprintf("algo:paged_index:%s:%d", abtest.App, params.UserId)
}

func (self *PagedRichStrategy) BuildData() error {
	params := self.ctx.GetRequest()
	if params.Offset > 0 {		// 只有非第一页才获取缓存
		self.cacheModule.GetStruct(self.CacheKey(), &self.pageIdsMap)
	}
	return nil
}

func (self *PagedRichStrategy) Strategy() error {
	for index := 0; index < self.ctx.GetDataLength(); index++ {
		dataInfo := self.ctx.GetDataByIndex(index)
		rankInfo := dataInfo.GetRankInfo()
		dataId := dataInfo.GetDataId()
		if value, ok := self.pageIdsMap[dataId]; ok {
			rankInfo.PagedIndex = value
		} else {
			rankInfo.PagedIndex = 9999999
		}
	}
	return nil
}

func (self *PagedRichStrategy) Logger() error {
	if response := self.ctx.GetResponse(); response != nil {
		for _, item := range response.DataList {
			self.pageIdsMap[item.DataId] = item.Index
		}
	}
	
	if len(self.pageIdsMap) > 0 {
		go func(){
			self.cacheModule.SetStruct(self.CacheKey(), self.pageIdsMap, 60 * 60, 0)
		}()
	}
	return nil
}

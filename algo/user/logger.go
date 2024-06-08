package user

import (
	"rela_recommend/algo"
	"rela_recommend/log"
	"rela_recommend/rpc/search"
)

type DoNearbySeenSearchLogger struct{}

// 已读接口调用
func (lg *DoNearbySeenSearchLogger) Do(ctx algo.IContext) error {
	response := ctx.GetResponse()
	seenIds := make([]int64, 0)
	if response != nil {
		for _, item := range response.DataList {
			seenIds = append(seenIds, item.DataId)
		}
	}
	if len(seenIds) > 0 {
		go func() {
			ok := search.CallNearbySeenList(ctx.GetRequest().UserId, "quick_match", seenIds)
			if !ok {
				log.Warn("search seen failed\n")
			}
		}()
	}
	return nil
}

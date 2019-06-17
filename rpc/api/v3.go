package api

import (
	"fmt"
	"errors"
	"rela_recommend/factory"
)
const internalBackendRecommendMomentListUrl = "/backend/moments/recommend"

type backendRecommendMomentDataRes struct {
	TopTheme        int64						`json:"topTheme"`
	RecommandThemes map[int64]int				`json:"recommandThemes"`
}

type backendRecommendMomentListRes struct {
	Code		int								`json:"code"`
	Message		string							`json:"message"`
	TTL			int								`json:"ttl"`
	Data		backendRecommendMomentDataRes	`json:"data"`
}

// 获取运营配置的日志置顶和推荐列表：1 为话题 2 为日志； return topMap, recommendMap
func CallBackendRecommendMomentList(category int) (map[int64]int, map[int64]int, error) {
	params := fmt.Sprintf("category=%d", category)
	res := &backendRecommendMomentListRes{}
	var topMap, recMap = map[int64]int{}, map[int64]int{}
	var err = factory.V3RpcClient.SendGETForm(internalBackendRecommendMomentListUrl, params, res)
	if err == nil {
		if res.Code == 0 {
			topMap[res.Data.TopTheme] = 1
			if res.Data.RecommandThemes != nil {
				recMap = res.Data.RecommandThemes
			}
		} else {
			err = errors.New(fmt.Sprintf("code:%d, message:%s", res.Code, res.Message))
		}
	}
	return topMap, recMap, err
}

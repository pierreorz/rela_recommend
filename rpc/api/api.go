package api

import (
	"fmt"
	"errors"
	"rela_recommend/factory"
)
const internalBackendRecommendMomentListUrl = "/internal/backend/moments/recommend"

type backendRecommendMomentDataRes struct {
	TopTheme        int64						`json:"topTheme"`
	RecommandThemes map[int64]int				`json:"recommendThemes"`
}

type backendRecommendMomentListRes struct {
	Code		int								`json:"code"`
	Message		string							`json:"message"`
	TTL			int								`json:"ttl"`
	Data		backendRecommendMomentDataRes	`json:"data"`
}

// 获取运营配置的日志置顶和推荐列表：1 为话题 2 为日志； return ids, topMap, recommendMap
func CallBackendRecommendMomentList(category int) ([]int64, map[int64]int, map[int64]int, error) {
	params := fmt.Sprintf("category=%d", category)
	res := &backendRecommendMomentListRes{}
	var ids, topMap, recMap = []int64{}, map[int64]int{}, map[int64]int{}
	var err = factory.ApiRpcClient.SendGETForm(internalBackendRecommendMomentListUrl, params, res)
	if err == nil {
		if res.Code == 0 {
			if res.Data.TopTheme > 0 {
				topMap[res.Data.TopTheme] = 1
				ids = append(ids, res.Data.TopTheme)
			}
			if res.Data.RecommandThemes != nil {
				recMap = res.Data.RecommandThemes
				for k, _ := range recMap {
					ids = append(ids, k)
				}
			}
		} else {
			err = errors.New(fmt.Sprintf("code:%d, message:%s", res.Code, res.Message))
		}
	}
	return ids, topMap, recMap, err
}


//********************************************************************************* 获取用户会员状态：level, error

const internalUserVipUrl = "/internal/users/vipSet/detail"

type userVipStatusDataRes struct {
	Level       int						`json:"level"`
	VipIcon		int8					`json:"vipIcon"`
}

type userVipStatusRes struct {
	Data        userVipStatusDataRes		`json:"data"`
}
func CallUserVipStatus(userId int64) (userVipStatusDataRes, error) {
	params := fmt.Sprintf("id=%d", userId)
	res := &userVipStatusRes{}
	var err = factory.ApiRpcClient.SendGETForm(internalUserVipUrl, params, res)
	if err == nil {
		return res.Data, nil
	}
	return res.Data, err
}

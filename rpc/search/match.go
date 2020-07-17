package search

import (
	"encoding/json"
	"fmt"
	"rela_recommend/algo"
	"rela_recommend/factory"
	"rela_recommend/log"
	"rela_recommend/models/redis"
	"strings"
)

// 搜索接口
const internalSearchMatchListUrl = "/search/quick_match"

type SearchMatchResDataItem struct {
	Id int64 `json:"id"`
}

type matchListResIds struct {
	Data      []SearchMatchResDataItem `json:"result_data"`
	TotalSize int64                    `json:"total_size"`
	AggsData  map[string]interface{}   `json:"-"`
	Result    string                   `json:"result"`
	ErrCode   string                   `json:"errcode"`
	ErrDesc   string                   `json:"errdesc"`
	ErrDescEn string                   `json:"errdesc_en"`
	ReqeustID string                   `json:"request_id"`
}

type searchMatchRequest struct {
	UserID    int64   `json:"userId" form:"userId"`
	Lng       float32 `json:"lng" form:"lng" `
	Lat       float32 `json:"lat" form:"lat" `
	PinnedIds string  `json:"pinned_ids" form:"pinned_ids" `
	Filter    string  `json:"filter" form:"filter" `
}

// 已读接口
const internalSearchMatchSeenListUrl = "/seen/quick_match"

type matchSeenListResIds struct {
	Data      string `json:"data"`
	Result    string `json:"result"`
	ErrCode   string `json:"errcode"`
	ErrDesc   string `json:"errdesc"`
	ErrDescEn string `json:"errdesc_en"`
	ReqeustID string `json:"request_id"`
}

type searchMatchSeenRequest struct {
	UserID     int64  `json:"userId" form:"userId"`
	Expiration int64  `json:"expiration" form:"expiration" `
	Scenery    string `json:"scenery" form:"scenery" `
	SeenIds    string `json:"seen_ids" form:"seen_ids" `
}

// 获取用户列表, 过滤条件：
// role_name = "1,2,3"
func CallMatchList(ctx algo.IContext, userId int64, lat, lng float32, userIds []int64, user *redis.UserProfile) ([]int64, error) {
	abtest := ctx.GetAbTest()
	idlist := make([]int64, 0)

	strIds := make([]string, len(userIds))
	for k, v := range userIds {
		strIds[k] = fmt.Sprintf("%d", v)
	}
	strsIds := strings.Join(strIds, ",")

	filters := []string{}
	if abtest.GetBool("filter_role_name", false) {
		if user != nil && user.WantRole != "" && user.WantRole != "0" {
			filters = []string{
				fmt.Sprintf("role_name:%s", user.WantRole),
			}
		}
	}

	params := searchMatchRequest{
		UserID:    userId,
		Lng:       lng,
		Lat:       lat,
		PinnedIds: strsIds,
		Filter:    strings.Join(filters, "*"),
	}
	if paramsData, err := json.Marshal(params); err == nil {
		res := &matchListResIds{}
		if err = factory.AiSearchRpcClient.SendPOSTJson(internalSearchMatchListUrl, paramsData, res); err == nil {
			for _, element := range res.Data {
				idlist = append(idlist, element.Id)
			}
			log.Infof("get paramsData:%s", string(paramsData))
			return idlist, err
		} else {
			return idlist, err
		}
	} else {
		return idlist, err
	}

}

// 获取已读用户列表
func CallMatchSeenList(userId, expiration int64, scenery string, userIds []int64) bool {

	strIds := make([]string, len(userIds))
	for k, v := range userIds {
		strIds[k] = fmt.Sprintf("%d", v)
	}
	strsIds := strings.Join(strIds, ",")

	params := searchMatchSeenRequest{
		UserID:     userId,
		Expiration: expiration,
		Scenery:    scenery,
		SeenIds:    strsIds,
	}
	if paramsData, err := json.Marshal(params); err == nil {
		log.Infof("paramsData%s", string(paramsData))
		res := &matchSeenListResIds{}
		if err = factory.AiSearchRpcClient.SendPOSTJson(internalSearchMatchSeenListUrl, paramsData, res); err == nil {
			return res.Data == "ok"
		} else {
			return false
		}
	} else {
		return false
	}

}

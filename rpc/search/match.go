package search

import (
	"encoding/json"
	"fmt"
	"rela_recommend/factory"
	"strings"
)

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
}

type searchMatchRequest struct {
	UserID    int64   `json:"userId" form:"userId"`
	Lng       float32 `json:"lng" form:"lng" `
	Lat       float32 `json:"lat" form:"lat" `
	PinnedIds string  `json:"pinned_ids" form:"pinned_ids" `
}

// 获取用户列表
func CallMatchList(userId int64, lat, lng float32, userIds []int64) ([]int64, error) {
	idlist := make([]int64, 0)

	strIds := make([]string, len(userIds))
	for k, v := range userIds {
		strIds[k] = fmt.Sprintf("%d", v)
	}
	strsIds := strings.Join(strIds, ",")

	params := searchMatchRequest{
		UserID:    userId,
		Lng:       lng,
		Lat:       lat,
		PinnedIds: strsIds,
	}
	if paramsData, err := json.Marshal(params); err == nil {
		res := &matchListResIds{}
		if err = factory.AiSearchRpcClient.SendPOSTJson(internalSearchMatchListUrl, paramsData, res); err == nil {
			for _, element := range res.Data {
				idlist = append(idlist, element.Id)
			}
			return idlist, err
		} else {
			return idlist, err
		}
	} else {
		return idlist, err
	}

}

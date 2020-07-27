package search

import (
	"fmt"
	"strings"
	"encoding/json"
	"rela_recommend/factory"
)

const internalSearchNearMomentListUrlV1 = "/search/friend_moments"

type SearchMomentResDataItem struct {
	Id int64 `json:"id"`
}

type searchMomentRes struct {
	Data      []SearchMomentResDataItem `json:"result_data"`
	TotalSize int                   `json:"total_size"`
	ErrCode   string                `json:"errcode"`
	ErrEsc    string                `json:"erresc"`
}

type searchMomentRequest struct {
	UserID    int64   `json:"userId" form:"userId"`
	Offset        int64   `json:"offset" form:"offset"`
	Limit         int64   `json:"limit" form:"limit"`
	Distance         string   `json:"distance" form:"distance"`
	Lng       float32 `json:"lng" form:"lng" `
	Lat       float32 `json:"lat" form:"lat" `
	Filter    string  `json:"filter" form:"filter" `
}

// 获取附近日志列表
func CallNearMomentListV1(userId int64, lat, lng float32, offset, limit int64, momentTypes string, insertTimestamp float32, distance string) ([]int64, error) {
	idlist := make([]int64, 0)
	filters := []string{
		fmt.Sprintf("{moments_type:%s}", momentTypes),                     //  moments Type
		fmt.Sprintf("insert_time:[%f,)", insertTimestamp),           // time
	}
	params:=searchMomentRequest{
		UserID:userId,
		Offset:offset,
		Limit:limit,
		Distance:distance,
		Lng:lng,
		Lat:lat,
		Filter:        strings.Join(filters, "*"),
	}

	if paramsData, err := json.Marshal(params); err == nil {
		searchRes := &searchMomentRes{}
		if err = factory.AiSearchRpcClient.SendPOSTJson(internalSearchNearMomentListUrlV1, paramsData, searchRes); err == nil {
			for _, element := range searchRes.Data {
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


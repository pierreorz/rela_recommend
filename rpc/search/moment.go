package search

import (
	"encoding/json"
	"fmt"
	"rela_recommend/factory"
	"rela_recommend/utils"
	"strings"
)

const internalSearchNearMomentListUrlV1 = "/search/friend_moments"

type SearchMomentResDataItem struct {
	Id int64 `json:"id"`
}

type searchMomentRes struct {
	Data      []SearchMomentResDataItem `json:"result_data"`
	TotalSize int                       `json:"total_size"`
	ErrCode   string                    `json:"errcode"`
	ErrEsc    string                    `json:"erresc"`
}

type searchMomentRequest struct {
	UserID   int64   `json:"userId" form:"userId"`
	Offset   int64   `json:"offset" form:"offset"`
	Limit    int64   `json:"limit" form:"limit"`
	Distance string  `json:"distance" form:"distance"`
	Lng      float32 `json:"lng" form:"lng" `
	Lat      float32 `json:"lat" form:"lat" `
	Filter   string  `json:"filter" form:"filter" `
}

// 获取附近日志列表
func CallNearMomentListV1(userId int64, lat, lng float32, offset, limit int64, momentTypes string, insertTimestamp float32, distance string) ([]int64, error) {
	idlist := make([]int64, 0)
	filters := []string{
		fmt.Sprintf("{moments_type:%s}", momentTypes),     //  moments Type
		fmt.Sprintf("insert_time:[%f,)", insertTimestamp), // time
	}
	params := searchMomentRequest{
		UserID:   userId,
		Offset:   offset,
		Limit:    limit,
		Distance: distance,
		Lng:      lng,
		Lat:      lat,
		Filter:   strings.Join(filters, "*"),
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

///////////////////////////////////////////  话题过滤+审核+推荐置顶

const internalSearchAuditUrl = "/search/audit_moment"

type SearchMomentAuditResDataItem struct {
	Id          int64  `json:"id"`
	TopType     string `json:"top_type"`
	AuditStatus int    `json:"audit_status"`
}

type searchMomentAuditRes struct {
	Data      []SearchMomentAuditResDataItem `json:"result_data"`
	TotalSize int                            `json:"total_size"`
	ErrCode   string                         `json:"errcode"`
	ErrEsc    string                         `json:"erresc"`
}

type searchMomentAuditRequest struct {
	UserID       int64  `json:"userId" form:"userId"`
	Filter       string `json:"filter" form:"filter" `
	ReturnFields string `json:"return_fields" form:"return_fields"`
}

// 获取附近日志列表
func CallMomentAuditMap(userId int64, moments []int64) ([]int64, map[int64]SearchMomentAuditResDataItem, error) {
	filters := []string{
		fmt.Sprintf("id:%s", utils.JoinInt64s(moments, ",")),
	}
	params := searchMomentAuditRequest{
		UserID:       userId,
		Filter:       strings.Join(filters, "*"),
		ReturnFields: "top_type,audit_status",
	}

	resIds := []int64{}
	resMap := map[int64]SearchMomentAuditResDataItem{}
	if paramsData, err := json.Marshal(params); err == nil {
		searchRes := &searchMomentAuditRes{}
		if err = factory.AiSearchRpcClient.SendPOSTJson(internalSearchAuditUrl, paramsData, searchRes); err == nil {
			for i, element := range searchRes.Data {
				resMap[element.Id] = searchRes.Data[i]
				resIds = append(resIds, element.Id)
			}
			return resIds, resMap, err
		} else {
			return resIds, resMap, err
		}
	} else {
		return resIds, resMap, err
	}
}

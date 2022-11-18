package search

import (
	"encoding/json"
	"fmt"
	"rela_recommend/factory"
	"strings"
)

const internalSearchLabelMomentsListUrl = "/search/label_moments"
const internalSearchLabelSuggestListUrl = "/search/label_suggest"
const internalSearchLabelListUrl = "/search/label"

type SearchLabelMomentResDataItem struct {
	Id int64 `json:"id"`
}

type searchLabelMomentRes struct {
	Data      []LabelResDataItem `json:"result_data"`
	TotalSize int                `json:"total_size"`
	ErrCode   string             `json:"errcode"`
	ErrEsc    string             `json:"erresc"`
}

type LabelResDataItem struct {
	Id          int64  `json:"id"`
	ViewNum     int64  `json:"view_num"`
	Name        string `json:"name"`
	StatJoinNum int64  `json:"stat_join_num"`
}

type searchLabelRes struct {
	Data      []LabelResDataItem `json:"result_data"`
	TotalSize int                `json:"total_size"`
	ErrCode   string             `json:"errcode"`
	ErrEsc    string             `json:"erresc"`
}

//获取标签下日志列表
func CallLabelMomentList(id int64, limit int64) ([]int64, map[int64]LabelResDataItem, error) {
	idList := make([]int64, 0)
	dataList := make(map[int64]LabelResDataItem, 0)
	filters := []string{
		fmt.Sprintf("main_id:%d", id), //  moments Type
	}

	params := searchBaseRequest{
		Filter:       strings.Join(filters, "*"),
		Limit:        limit,
		Sort:         "-insert_time",
		ReturnFields: "stat_join_num",
	}

	if paramsData, err := json.Marshal(params); err == nil {
		searchRes := &searchLabelMomentRes{}
		if err = factory.MomentSearchRpcClient.SendPOSTJson(internalSearchLabelMomentsListUrl, paramsData, searchRes); err == nil {
			for _, item := range searchRes.Data {
				idList = append(idList, item.Id)
				dataList[item.Id] = item
			}
			return idList, dataList, err
		} else {
			return idList, dataList, err
		}
	} else {
		return idList, dataList, err
	}
}

//标签联想

func CallLabelSuggestList(query string) ([]int64, error) {
	namelist := make([]int64, 0)

	params := searchBaseRequest{
		Query: query,
	}

	if paramsData, err := json.Marshal(params); err == nil {
		searchRes := &searchLabelRes{}
		if err = factory.MomentSearchRpcClient.SendPOSTJson(internalSearchLabelSuggestListUrl, paramsData, searchRes); err == nil {
			for _, element := range searchRes.Data {
				namelist = append(namelist, element.Id)
			}
			return namelist, err
		} else {
			return namelist, err
		}
	} else {
		return namelist, err
	}
}

//标签搜索接口
func CallLabelSearchList(query string, limit int64) ([]int64, error) {
	namelist := make([]int64, 0)

	params := searchBaseRequest{
		Query: query,
		Limit: limit,
	}

	if paramsData, err := json.Marshal(params); err == nil {
		searchRes := &searchLabelRes{}
		if err = factory.MomentSearchRpcClient.SendPOSTJson(internalSearchLabelListUrl, paramsData, searchRes); err == nil {
			for _, element := range searchRes.Data {
				namelist = append(namelist, element.Id)
			}
			return namelist, err
		} else {
			return namelist, err
		}
	} else {
		return namelist, err
	}
}

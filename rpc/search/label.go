package search

import (
	"encoding/json"
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
	Data      []SearchLabelMomentResDataItem `json:"result_data"`
	TotalSize int                       `json:"total_size"`
	ErrCode   string                    `json:"errcode"`
	ErrEsc    string                    `json:"erresc"`
}



type SearchLabelResDataItem struct {
	ViewNum   int64 `json:"view_num"`
	Name      string `json:"name"`
	JoinNum    int64 `json:"join_num"`
}

type searchLabelRes struct {
	Data      []SearchLabelResDataItem `json:"result_data"`
	TotalSize int                       `json:"total_size"`
	ErrCode   string                    `json:"errcode"`
	ErrEsc    string                    `json:"erresc"`
}


//获取广告日志列表
func CallLabelMomentList(id int64) ([]int64, error) {
	idlist := make([]int64, 0)
	filters := []string{
		fmt.Sprintf("{moments_type:%s}", "ad"),     //  moments Type
		fmt.Sprintf("ad_location.start_time:(,now/m]", ), // time
		fmt.Sprint("ad_location.end_time:[now/m,)"),
	}

	params := searchMomentRequest{
		UserID:   userId,
		Filter:   strings.Join(filters, "*"),
	}

	if paramsData, err := json.Marshal(params); err == nil {
		searchRes := &searchMomentRes{}
		if err = factory.MomentSearchRpcClient.SendPOSTJson(internalSearchNearMomentListUrlV1, paramsData, searchRes); err == nil {
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
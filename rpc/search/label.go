package search

import (
	"encoding/json"
	"fmt"
	"rela_recommend/factory"
	"strings"
)

const internalSearchLabelMomentListUrl = "/search/label_monents"
const internalSearchLabelSuggestListUrl = "/search/label_suggest"
const internalSearchLabelListUrl = "/search/label"


type searchBaseLabelRequest struct {
	// 请求者 ID
	UserID       int64   `json:"userId" form:"userId"`
	// 翻页起始点
	Offset       int64   `json:"offset" form:"offset"`
	// 每页个数
	Limit        int64   `json:"limit" form:"limit"`
	// 搜索词
	Query        string `form:"query" json:"query"`
	// 过滤条件，有一定语法规则，见下
	Filter       string  `json:"filter" form:"filter" `
	// 按字段排序，相当于 sql order by，有一定语法规则，见下
	Sort         string `form:"sort" json:"sort"`
	// 返回字段，多字段按逗号分隔，相当于 sql select
	ReturnFields string `form:"return_fields" json:"return_fields"`
}


type searchMomentResponse struct {
	// 搜索结果，不同场景可以有不同的 resDataItem 定义，如 user, moment
	Data      []resDataItem          `json:"result_data"`
	// 总数，用于计算分页
	TotalSize int64                  `json:"total_size"`
	// 错误码，正常结果 errcode == "0"， 空结果 errcode == "-1"，其他错误样例 errcode == "param_error"
	ErrCode   string                 `json:"errcode"`
	// 错误描述
	ErrDesc   string                 `json:"errdesc"`
	// 请求ID
	ReqeustID string                 `json:"request_id"`
}

type resDataItem struct {
	Id int64 `json:"id"`
}


type searchMomentLabelResponse struct {
	// 搜索结果，不同场景可以有不同的 resDataItem 定义，如 user, moment
	Data      []resDataItemv1          `json:"result_data"`
	// 总数，用于计算分页
	TotalSize int64                  `json:"total_size"`
	// 错误码，正常结果 errcode == "0"， 空结果 errcode == "-1"，其他错误样例 errcode == "param_error"
	ErrCode   string                 `json:"errcode"`
	// 错误描述
	ErrDesc   string                 `json:"errdesc"`
	// 请求ID
	//ReqeustID string                 `json:"request_id"`
}


type resDataItemv1 struct {
	ViewNum int64 `json:"view_num"`
	Name    string `json:"name"`
	JoinNum  int64   `json:"join_num"`
	Id        int64   `json:"id"`
	Status    int      `json:"status"`
}


func CallLabelMomentList(id int64) ([]int64, error) {
	idlist := make([]int64, 0)

	filters := []string{
		fmt.Sprintf("user_id:%s", strings.Join(userIdListstr, ",")),
	}
	params := searchBaseLabelRequest{
		Filter: strings.Join(filters, "*"),
	}

	if paramsData, err := json.Marshal(params); err == nil {
		searchRes := &searchMomentRes{}
		if err = factory.MomentSearchRpcClient.SendPOSTJson(internalSearchLiveMomentListUrl, paramsData, searchRes); err == nil {
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

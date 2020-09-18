package search

import (
	"encoding/json"
	"rela_recommend/factory"
	"strings"
)

type SearchResDataItem struct {
	Id int64 `json:"id"`
}

type searchBaseResponse struct {
	Data      []SearchResDataItem `json:"result_data"`
	TotalSize int                 `json:"total_size"`
	ErrCode   string              `json:"errcode"`
	ErrEsc    string              `json:"erresc"`
}

type searchBaseRequest struct {
	UserID   int64   `json:"userId" form:"userId"`
	Offset   int64   `json:"offset" form:"offset"`
	Limit    int64   `json:"limit" form:"limit"`
	Distance string  `json:"distance" form:"distance"`
	Lng      float32 `json:"lng" form:"lng" `
	Lat      float32 `json:"lat" form:"lat" `
	Filter   string  `json:"filter" form:"filter" `
	Query    string  `json:"query" form:"query" `
}

func CallSearchIdList(url string, userId int64, lat, lng float32, offset, limit int64, filters []string, query string) ([]int64, error) {
	idlist := make([]int64, 0)
	params := searchBaseRequest{
		UserID: userId,
		Offset: offset,
		Limit:  limit,
		Lng:    lng,
		Lat:    lat,
		Filter: strings.Join(filters, "*"),
		Query:  query,
	}

	if paramsData, err := json.Marshal(params); err == nil {
		searchRes := &searchBaseResponse{}
		if err = factory.AiSearchRpcClient.SendPOSTJson(url, paramsData, searchRes); err == nil {
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

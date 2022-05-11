package search

import (
	"encoding/json"
	"fmt"
	"rela_recommend/algo"
	"rela_recommend/factory"
	"rela_recommend/log"
	"strings"
	"time"
)

const internalSearchMateTextListUrl = "/search/mate_text"

type MateTextResDataItem struct {
	Id     int64         `json:"id"`
	Text   string        `json:"text"`
	Cities []interface{} `json:"cities"`
	Weight int           `json:"weight"`
}

type mateTextRes struct {
	Data      []MateTextResDataItem `json:"result_data"`
	TotalSize int                   `json:"total_size"`
	ErrCode   string                `json:"errcode"`
	ErrEsc    string                `json:"erresc"`
}

type mateSearchRequest struct {
	UserID        int64   `json:"userId" form:"userId"`
	Offset        int64   `json:"offset" form:"offset"`
	Limit         int64   `json:"limit" form:"limit"`
	Lng           float32 `json:"lng" form:"lng" `
	Lat           float32 `json:"lat" form:"lat" `
	MobileOS      string  `json:"mobileOS" form:"mobileOS"`
	ClientVersion int     `json:"clientVersion" form:"clientVersion"`
	Filter        string  `json:"filter" form:"filter" `
	ReturnFields  string  `json:"return_fields" form:"return_fields"`
	Distance      string  `json:"distance" form:"distance"`
}

func CallMateTextList(request *algo.RecommendRequest, searchLimit int64) ([]MateTextResDataItem, error) {
	localTimeStr, ok := request.Params["local_time"]
	if !ok {
		return nil, nil
	}

	localTime, err := time.Parse("2006-01-02 15:04:05", localTimeStr)
	if err != nil {
		return nil, err
	}

	localHour := localTime.Hour()

	log.Infof("mate local hour: %+v, %d", localTime, localHour)
	if localHour >= 0 && localHour <= 2 {
		localHour += 24 // 22点到凌晨2点会跨天，+24方便比较
	}
	filters := []string{
		fmt.Sprintf("start_hour:(,%d]*end_hour:[%d,)", localHour, localHour), // base
	}

	params := mateSearchRequest{
		UserID:        request.UserId,
		Offset:        request.Offset,
		Limit:         searchLimit,
		Lng:           request.Lng,
		Lat:           request.Lat,
		MobileOS:      request.MobileOS,
		ClientVersion: request.ClientVersion,
		Filter:        strings.Join(filters, "*"),
		ReturnFields:  "*",
		Distance:      "50km",
	}
	log.Infof("search=================%+v",params)
	if paramsData, err := json.Marshal(params); err == nil {
		searchRes := &mateTextRes{}
		if err = factory.AiSearchRpcClient.SendPOSTJson(internalSearchMateTextListUrl, paramsData, searchRes); err == nil {
			return searchRes.Data, err
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

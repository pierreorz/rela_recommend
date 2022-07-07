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
	Weight int           `json:"weight"`
	TextType int64      `json:"textType" `
	TagType  int64      `json:"tagType" `
	UserId int64 		`json:"userId"`
	ImageUrl string     `json:"imageUrl"`
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
	TextType      string  `json:"textType" form:"textType"`
	TagType 	  string  `json:"tagType" form:"tagType"`
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
		ReturnFields:  "id,tag_type,text_type,weight",
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

type SearchType struct {
	TextType string `json:"textType"`
	TagType  []string `json:"tagType"`
}

func CategMateTextList(request *algo.RecommendRequest, searchLimit int64, TagType string ,TextType []string ) ([]MateTextResDataItem, error) {
	//获取文案的文本类型和文案的状态

	// 增加时间过滤
	localTimeStr, ok := request.Params["local_time"]
	if !ok {
		return nil, nil
	}
	localTime, err := time.Parse("2006-01-02 15:04:05", localTimeStr)
	if err != nil {
		return nil, err
	}

	localHour := localTime.Hour()
	if localHour >= 0 && localHour <= 2 {
		localHour += 24 // 22点到凌晨2点会跨天，+24方便比较
	}
	filters := []string{}
	timeSenList:=[]string{"start_hour:(,8]*end_hour:[10,)","start_hour:(,12]*end_hour:[14,)","start_hour:(,15]*end_hour:[17,)","start_hour:(,22]*end_hour:[2,)"}
	var timeSen string
	var hotSen string
	if localHour >= 8 && localHour <= 10 {
		timeSen=timeSenList[0]+"*text_type:30"
	}
	if localHour >= 12 && localHour <= 14 {
		timeSen=timeSenList[1]+"*text_type:30"
	}
	if localHour >= 15 && localHour <= 17 {
		timeSen=timeSenList[2]+"*text_type:30"
	}
	if localHour >= 22 && localHour <= 2 {
		timeSen=timeSenList[3]+"*text_type:30"
	}
	if localHour >= 20 && localHour <= 23 {
		hotSen = timeSenList[4] + "*text_type:30"
	}
	filters=append(filters,timeSen)
	if len(hotSen)!=0{
		filters=append(filters,hotSen)
	}

	tagLine:=TagType+"*{"+strings.Join(TextType, "|")+"}"
	filters=append(filters, tagLine)

	params := mateSearchRequest{
		UserID:        request.UserId,
		Offset:        request.Offset,
		Limit:         searchLimit,
		Lng:           request.Lng,
		Lat:           request.Lat,
		MobileOS:      request.MobileOS,
		Filter:        strings.Join(filters, "|"),
		ReturnFields:  "*",
		Distance:      "50km",

	}
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



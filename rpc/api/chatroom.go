package api

import (
	"errors"
	"fmt"
	"rela_recommend/factory"
	"time"
)

const internalSearchChatRoomListUrl = "/internal/chatrooms"

// 接口调用参数 -1.all; 1. video; 2. audio; 3. multi_audio(radio)
var ChatRoomLiveTypes []int = []int{-1, 1, 2, 3}

type SimpleChatroom struct {
	UserID           int64     `json:"uid"`
	Lat              float32   `json:"lat"`
	Lng              float32   `json:"lng"`
	GemProfit        float32   `json:"gemProfit"`
	LiveType         int       `json:"liveType"`
	SendMsgCount     int       `json:"sendMsgCount"`
	ReceivedMsgCount int       `json:"receivedMsgCount"`
	ShareCount       int       `json:"shareCount"`
	Score            float32   `json:"score"`
	BottomScore      int       `json:"bottomScore"`
	FansCount        int       `json:"fansCount"`
	Priority         float32   `json:"priority"`
	Recommend        int       `json:"recommend"`
	RecommendLevel   int       `json:"recommendLevel"`
	StarsCount       int       `json:"starsCount"`
	TopCount         int       `json:"topCount"`
	TopView          int       `json:"topView"`
	NowIncoming      float32    `json:"nowGem"`
	DayIncoming      float32   `json:"dayIncoming"`
	MonthIncoming    float32   `json:"monthIncoming"`
	IsMulti          int       `json:"isMulti"`
	Classify         int       `json:"classify"`
	MomentsID        int64     `json:"momentsId"`
	CreateTime       time.Time `json:"createTime"`
	IsShowAdd        int        `json:"is_show_add"`
	Data             string    `json:"data"`
}

type ChatRoomRes struct {
	Code    int              `json:"code"`
	Message string           `json:"message"`
	TTL     int              `json:"ttl"`
	Data    []SimpleChatroom `json:"data"`
}

// 获取直播列表
func CallChatRoomList(liveType int) ([]SimpleChatroom, error) {
	params := fmt.Sprintf("type=%d", liveType)
	res := &ChatRoomRes{}
	err := factory.ChatRoomRpcClient.SendGETForm(internalSearchChatRoomListUrl, params, res)
	if err == nil {
		if res != nil && res.Code == 0 && res.Data != nil {
			return res.Data, nil
		} else {
			errMsg := fmt.Sprintf("result error, %+v", res)
			return nil, errors.New(errMsg)
		}
	} else {
		return nil, err
	}
}

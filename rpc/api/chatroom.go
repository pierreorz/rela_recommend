package api

import (
	"errors"
	"fmt"
	"rela_recommend/factory"
	"time"
)

const (
	FREE = iota

	LinkMicWait // wait for link mic
	LinkMicBusy // linking

	PkWait    // wait for pk
	PkBusy    // pking
	PkSummary // pk summary

	MultiAudio          // multi link
	MultiAudioEncounter // multi link and encounter
	MultiVideoFour
	MultiVideoNine
	Blind

	internalSearchChatRoomListUrl = "/internal/chatrooms"
)
const (
	// HourRankLabelWeight https://wiki.rela.me/pages/viewpage.action?pageId=30474709
	// 优先比较 level , level 相同则比较 weight
	HourRankLabelWeight = iota
	RecommendLabelWeight
	WeekStarLabelWeight
	MonthStarLabelWeight
	HoroscopeLabelWeight
	ModalStudentLabelWeight
	TypeLabelWeight
	ClassifyLabelWeight
	AroundWeight
	CityWeight
	FollowLabelWeight
	StrategyLabelWeight

	TypeRecommend     = 1
	TypeBigVideo      = 32768
	TypeBigMultiAudio = 65535
	TypeGroupVideo    = 65536
)

// ChatRoomLiveTypes 接口调用参数 -1.all; 1. video; 2. audio; 3. multi_audio(radio) 4. group video
var ChatRoomLiveTypes = []int{-1, 1, 2, 3, 4}

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
	NowIncoming      float32   `json:"nowGem"`
	DayIncoming      float32   `json:"dayIncoming"`
	MonthIncoming    float32   `json:"monthIncoming"`
	IsMulti          int       `json:"isMulti"`
	Classify         int       `json:"classify"`
	MomentsID        int64     `json:"momentsId"`
	CreateTime       time.Time `json:"createTime"`
	IsShowAdd        int       `json:"is_show_add"`
	Data             string    `json:"data"`
}

type ChatRoomRes struct {
	Code    int              `json:"code"`
	Message string           `json:"message"`
	TTL     int              `json:"ttl"`
	Data    []SimpleChatroom `json:"data"`
}

type ILiveRankItemV3 struct {
	Rank      int            `json:"rank"`      //等级
	Score     int            `json:"score"`     //观看人数
	Label     string         `json:"label"`     //推荐标签
	Recommend int            `json:"recommend"` //推荐类型
	LiveId    int64          `json:"liveID"`    //直播ID
	UserId    int64          `json:"user_id"`   //主播ID
	Status    int            `json:"status"`    //直播间状态
	Classify  int            `json:"classify"`  //直播分类
	Seats     []SeatInfo     `json:"seats"`
	LabelLang *multiLanguage `json:"labelLang"`
}

func (lrt *ILiveRankItemV3) GetLiveType() int {
	switch lrt.Status {
	case LinkMicBusy:
		return 1
	case PkBusy, PkSummary:
		return 2
	case MultiVideoFour, MultiVideoNine:
		return 3
	case Blind:
		return 4
	}
	return 0
}

// IsGroupVideo 四人、九人叽叽喳喳都是群播
func (lrt *ILiveRankItemV3) IsGroupVideo() bool {
	switch lrt.Status {
	case MultiVideoFour, MultiVideoNine:
		return true
	}
	return false
}

type SeatInfo struct {
	MicStatus     string `json:"micStatus"`
	SeatStatus    string `json:"seatStatus"`
	HeartNum      int    `json:"heartNum"`
	Choice        int    `json:"choice"`
	LastMicStatus string `json:"lastMicStatus"`
	BaseUserInfo
}

type BaseUserInfo struct {
	UserId   string `json:"userId"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

type multiLanguage struct {
	Chs string `json:"chs"`
	Cht string `json:"cht"`
	En  string `json:"en"`
	Url string `json:"url"`
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

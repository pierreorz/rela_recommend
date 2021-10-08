package live

import (
	"encoding/json"
	"rela_recommend/algo"
	"rela_recommend/algo/utils"
	"rela_recommend/log"
	"rela_recommend/models/behavior"
	"rela_recommend/models/pika"
	"rela_recommend/models/redis"
	rutils "rela_recommend/utils"
	"sort"
)

const (
	HourRankLabelWeight = iota
	RecommendLabelWeight
	WeekStarLabelWeight
	LiveTypeLabelWeight
	ClassifyLabelWeight

	HourRankLabel  = 1
	RecommendLabel = 2
	WeekStarLabel  = 3
	PkLabel        = 4
	BeamingLabel   = 5
	ClassifyLabel  = 6
)

// 用户信息
type UserInfo struct {
	UserId       int64
	UserCache    *redis.UserProfile
	LiveProfile  *redis.LiveProfile
	UserConcerns *rutils.SetInt64
}

func (self *UserInfo) GetBehavior() *behavior.UserBehavior {
	return nil
}

type LiveData struct {
	PreHourIndex int // 小时榜排名，1开始
	PreHourRank  int // 小时榜排名，1开始, 相同分数有并列名次
	labelList    []*labelItem
}

func (ld *LiveData) AppendLabelList(item *labelItem) {
	if ld.labelList == nil {
		ld.labelList = make([]*labelItem, 0)
	}

	ld.labelList = append(ld.labelList, item)
	sort.SliceStable(ld.labelList, func(i, j int) bool {
		iItem := ld.labelList[i]
		jItem := ld.labelList[j]
		return iItem.weight <= jItem.weight
	})

	if len(ld.labelList) > 2 {
		ld.labelList = ld.labelList[:2]
	}
}

// 主播信息
type LiveInfo struct {
	UserId           int64
	UserCache        *redis.UserProfile
	LiveProfile      *redis.LiveProfile
	LiveCache        *pika.LiveCache
	LiveData         *LiveData
	RankInfo         *algo.RankInfo
	Features         *utils.Features
	UserItemBehavior *behavior.UserBehavior
}

type labelItem struct {
	Title multiLanguage `json:"title"`
	Style int           `json:"style"`

	weight int
}

type multiLanguage struct {
	Chs string `json:"chs"`
	Cht string `json:"cht"`
	En  string `json:"en"`
}

type ILiveRankItemV3 struct {
	Rank      int          `json:"rank"`      //等级
	Score     int          `json:"score"`     //观看人数
	Label     string       `json:"label"`     //推荐标签
	Recommend int          `json:"recommend"` //推荐类型
	LiveId    int64        `json:"liveID"`    //直播ID
	UserId    int64        `json:"user_id"`   //主播ID
	Status    int          `json:"status"`    //直播间状态
	Classify  int          `json:"classify"`  //直播分类
	Seats     []SeatInfo   `json:"seats"`
	LabelList []*labelItem `json:"label_list"` //排序好的直播标签，见 https://wiki.rela.me/pages/viewpage.action?pageId=30474709
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

func (self *LiveInfo) GetDataId() int64 {
	return self.UserId
}

func (self *LiveInfo) GetResponseData(ctx algo.IContext) interface{} {

	if self.LiveCache != nil {
		var data ILiveRankItemV3
		err := json.Unmarshal([]byte(self.LiveCache.Data4Api.(string)), &data)
		if err != nil {
			log.Errorf("unmarshal live data %+v error: %+v", self.LiveCache.Data4Api, err)
			return nil
		}
		if len(data.Label) > 0 {
			self.LiveData.AppendLabelList(&labelItem{
				Style: RecommendLabel,
				Title: multiLanguage{
					Chs: data.Label,
					Cht: "",
					En:  "",
				},
				weight: RecommendLabelWeight,
			})
		}

		switch data.Status {
		case 0:
			self.LiveData.AppendLabelList(&labelItem{
				Style: PkLabel,
				Title: multiLanguage{
					Chs: "PK中",
					Cht: "PK中",
					En:  "PK",
				},
				weight: LiveTypeLabelWeight,
			})
		case 1:
			self.LiveData.AppendLabelList(&labelItem{
				Style: BeamingLabel,
				Title: multiLanguage{
					Chs: "连麦中",
					Cht: "連麥中",
					En:  "Beaming",
				},
				weight: LiveTypeLabelWeight,
			})
		}

		return self.LiveCache.Data4Api
	} else {
		return nil
	}
}

func (self *LiveInfo) SetRankInfo(rankInfo *algo.RankInfo) {
	self.RankInfo = rankInfo
}

func (self *LiveInfo) GetRankInfo() *algo.RankInfo {
	return self.RankInfo
}

func (self *LiveInfo) GetBehavior() *behavior.UserBehavior {
	return nil
}

func (self *LiveInfo) GetUserBehavior() *behavior.UserBehavior {
	return nil
}

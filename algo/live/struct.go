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
	"time"
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

	typeRecommend     = 1
	typeBigVideo      = 32768
	typeBigMultiAudio = 65535
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
)

var classifyMap map[int]multiLanguage

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

type ClassifyItem struct {
	Id            int64     `gorm:"primary_key;column:id" json:"id"`
	Rank          int       `gorm:"column:rank" json:"rank,omitempty"`
	Icon          string    `gorm:"column:icon" json:"icon,omitempty"`
	SelectIcon    string    `gorm:"column:selected_icon" json:"selected_icon,omitempty"`
	Status        int       `gorm:"column:status" json:"status,omitempty"`
	IsRecommended int       `gorm:"column:is_recommended" json:"is_recommended,omitempty"`
	SkillsJson    string    `gorm:"column:skills" json:"skills,omitempty"`
	Guide         string    `gorm:"column:guide" json:"guide,omitempty"`
	OnlineTime    time.Time `gorm:"column:online_time" json:"online_time,omitempty"`
	OfflineTime   time.Time `gorm:"column:offline_time" json:"offline_time,omitempty"`

	skill multiLanguage
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
	LabelList []*labelItem   `json:"label_list"` //排序好的直播标签，见 https://wiki.rela.me/pages/viewpage.action?pageId=30474709
}

func (lrt *ILiveRankItemV3) GetLiveType() int {
	switch lrt.Status {
	case LinkMicBusy:
		return 1
	case PkBusy, PkSummary:
		return 2
	}
	return 0
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
	params := ctx.GetRequest()

	// 只有在“推荐、视频、热聊“的情况下，返回label_list
	var needReturnLabel bool
	classify := rutils.GetInt(params.Params["classify"])
	switch classify {
	case typeRecommend, typeBigVideo, typeBigMultiAudio:
		needReturnLabel = true
	}

	if needReturnLabel && self.LiveCache != nil {
		var data ILiveRankItemV3
		err := json.Unmarshal([]byte(self.LiveCache.Data4Api.(string)), &data)
		if err != nil {
			log.Errorf("unmarshal live data %+v error: %+v", self.LiveCache.Data4Api, err)
			return nil
		}
		if len(data.Label) > 0 && data.LabelLang != nil {
			self.LiveData.AppendLabelList(&labelItem{
				Style: RecommendLabel,
				Title: multiLanguage{
					Chs: data.LabelLang.Chs,
					Cht: data.LabelLang.Cht,
					En:  data.LabelLang.En,
				},
				weight: RecommendLabelWeight,
			})
		}

		if classifyMap != nil {
			if lang, ok := classifyMap[data.Classify]; ok {
				self.LiveData.AppendLabelList(&labelItem{
					Title:  lang,
					Style:  ClassifyLabel,
					weight: ClassifyLabelWeight,
				})
			}
		}

		switch data.GetLiveType() {
		case 2:
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

		data.LabelList = self.LiveData.labelList

		dataJson, err := json.Marshal(data)
		if err == nil {
			return string(dataJson)
		}
		log.Errorf("marshal live data %+v err: %+v", data, err)
		return nil
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

package live

import (
	"encoding/json"
	"rela_recommend/algo"
	"rela_recommend/algo/utils"
	"rela_recommend/factory"
	"rela_recommend/help"
	"rela_recommend/log"
	"rela_recommend/models/behavior"
	"rela_recommend/models/pika"
	"rela_recommend/models/redis"
	rutils "rela_recommend/utils"
	"strconv"
	"time"
)

const (
	// https://wiki.rela.me/pages/viewpage.action?pageId=30474709
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

	HourRankLabel  = 1
	RecommendLabel = 2
	WeekStarLabel  = 3
	PkLabel        = 4
	BeamingLabel   = 5
	ClassifyLabel  = 6
	StrategyLabel  = 6
	FollowLabel    = 6
	AroundLabel    = 6
	CityLabel      = 6
	MultiBeamingLabel   =7

	typeRecommend     = 1
	typeBigVideo      = 32768
	typeBigMultiAudio = 65535

	level1 = 1
	level2 = 2
	level3 = 3

	LabelExpire = 86400

	prefix = "liveLabelList"
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
)

var classifyMap map[int]multiLanguage

// 用户信息
type UserInfo struct {
	UserId        int64
	UserCache     *redis.UserProfile
	LiveProfile   *redis.LiveProfile
	UserConcerns  *rutils.SetInt64
	UserInterests *rutils.SetInt64
	ConsumeUser    int
}

func (self *UserInfo) GetBehavior() *behavior.UserBehavior {
	return nil
}

type LiveData struct {
	PreHourIndex int // 小时榜排名，1开始
	PreHourRank  int // 小时榜排名，1开始, 相同分数有并列名次
	level1Label  *labelItem
	level2Label  *labelItem
	level3Label  *labelItem
}

func (ld *LiveData) AddLabel(item *labelItem) {
	switch item.level {
	case level1:
		if ld.level1Label == nil {
			ld.level1Label = item
		} else if ld.level1Label.weight > item.weight {
			ld.level1Label = item
		}
	case level2:
		if ld.level2Label == nil {
			ld.level2Label = item
		} else if ld.level2Label.weight > item.weight {
			ld.level2Label = item
		}
	case level3:
		if ld.level3Label == nil {
			ld.level3Label = item
		} else if ld.level3Label.weight > item.weight {
			ld.level3Label = item
		}
	}
}

func (ld *LiveData) ToLabelList() []*labelItem {
	var labels []*labelItem
	for _, l := range []*labelItem{ld.level1Label, ld.level2Label, ld.level3Label} {
		if l != nil {
			labels = append(labels, l)
		}
		if len(labels) >= 2 {
			break
		}
	}
	return labels
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

type newStyle struct{
	Font string  `json:"font"`
	Background string `json:"background"`
	Color     string   `json:"color"`
}
type labelItem struct {
	Title multiLanguage `json:"title"`
	Style int           `json:"style"`
	NewStyle newStyle	`json:"new_style"`
	weight int
	level  int
}

type multiLanguage struct {
	Chs string `json:"chs"`
	Cht string `json:"cht"`
	En  string `json:"en"`
	Url string  `json:"url"`
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
	case MultiVideoFour, MultiVideoNine:
		return 3
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
	userId := params.UserId
	if self.LiveCache != nil {
		liveLabelSwitchON := ctx.GetAbTest().GetBool("live_label_switch", false)

		params := ctx.GetRequest()

		// 只有在“推荐、视频、热聊“的情况下，返回label_list
		var needReturnLabel bool
		classify := rutils.GetInt(params.Params["classify"])
		switch classify {
		case typeRecommend, typeBigVideo, typeBigMultiAudio:
			needReturnLabel = true
		}

		if liveLabelSwitchON && needReturnLabel {
			if dataStr, ok := self.LiveCache.Data4Api.(string); ok {

				var data ILiveRankItemV3
				err := json.Unmarshal([]byte(dataStr), &data)
				if err != nil {
					log.Errorf("unmarshal live data %+v error: %+v", self.LiveCache.Data4Api, err)
					return nil
				}
				if len(data.Label) > 0 && data.LabelLang != nil {
					self.LiveData.AddLabel(&labelItem{
						Style: RecommendLabel,
						NewStyle:newStyle{
							Font:       "",
							Background: data.LabelLang.Url,
							Color:      "313333",
						},
						Title: multiLanguage{
							Chs: data.LabelLang.Chs,
							Cht: data.LabelLang.Cht,
							En:  data.LabelLang.En,
						},
						weight: RecommendLabelWeight,
						level:  level1,
					})
				}

				if classifyMap != nil {
					if lang, ok := classifyMap[data.Classify]; ok {
						self.LiveData.AddLabel(&labelItem{
							Title:  lang,
							NewStyle:newStyle{
								Font:       "",
								Background: "https://static.rela.me/whitetag.jpg",
								Color:      "ffffff",
							},
							Style:  ClassifyLabel,
							weight: ClassifyLabelWeight,
							level:  level3,
						})
					}
				}

				switch data.GetLiveType() {
				case 3:
					self.LiveData.AddLabel(&labelItem{
						Style: MultiBeamingLabel,
						NewStyle:newStyle{
							Font:       "",
							Background: "https://static.rela.me/bluengreentag.jpg",
							Color:      "ffffff",
						},
						Title: multiLanguage{
							Chs: "姬姬喳喳",
							Cht: "姬姬喳喳",
							En:  "Group Video",
						},
						weight: TypeLabelWeight,
						level:  level2,
					})
				case 2:
					self.LiveData.AddLabel(&labelItem{
						Style: PkLabel,
						NewStyle:newStyle{
							Font:       "",
							Background: "https://static.rela.me/Go5pifQDN4LnBuZzE2NjE0NzkzNjk4NzY=.png",
							Color:      "ffffff",
						},
						Title: multiLanguage{
							Chs: "PK中",
							Cht: "PK中",
							En:  "PK",
						},
						weight: TypeLabelWeight,
						level:  level2,
					})
				case 1:
					self.LiveData.AddLabel(&labelItem{
						Style: BeamingLabel,
						NewStyle:newStyle{
							Font:       "",
							Background: "https://static.rela.me/bluengreentag.jpg",
							Color:      "ffffff",
						},
						Title: multiLanguage{
							Chs: "连麦中",
							Cht: "連麥中",
							En:  "Beaming",
						},
						weight: TypeLabelWeight,
						level:  level2,
					})
				}

				data.LabelList = self.LiveData.ToLabelList()
				key := prefix + ":" + strconv.FormatInt(self.UserId, 10) + ":" + strconv.FormatInt(userId, 10)
				log.Warnf("label list key,%s", key)
				if err := help.SetExStructByCache(factory.CacheCommonRds, key, data.LabelList, LabelExpire); err != nil {
					log.Warnf("read label list err %s", err)
				}
				dataJson, err := json.Marshal(data)
				if err == nil {
					return string(dataJson)
				}
				log.Errorf("marshal live data %+v err: %+v", data, err)
				return nil
			}
		} else {
			return self.LiveCache.Data4Api
		}
	} else {
		return nil
	}
	return nil
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

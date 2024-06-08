package pika

import (
	"encoding/json"
	"rela_recommend/cache"
	"rela_recommend/help"
	"rela_recommend/log"
	"rela_recommend/utils"
	"strings"
	"time"
)

type LiveProfile struct {
	LiveId           int64    `json:"id"` // 用户ID
	LiveTypeId       int      `json:"liveTypeId"`
	UserIdStr        string   `json:"userId"`
	UserId           int64    `json:"-"`
	Text             string   `json:"text"`
	CreateTime       JsonTime `json:"createTime"`
	UpdateTime       JsonTime `json:"updateTime"`
	Active           int      `json:"active"`
	TopView          int      `json:"topView"`
	Ip               string   `json:"ip"`
	Ua               string   `json:"ua"`
	Lat              float32  `json:"lat"`
	Lng              float32  `json:"lng"`
	City             string   `json:"city"`
	GemProfitStr     string   `json:"gemProfit"`
	GemProfit        float32  `json:"-"`
	SendMsgCount     int      `json:"sendMsgCount"`
	ReceivedMsgCount int      `json:"receivedMsgCount"`
	ShareCount       int      `json:"shareCount"`
	AudioType        int      `json:"audioType"`
	IsMulti          int      `json:"isMulti"`
	Classify         int      `json:"classify"`
	MomentsID        int64    `json:"momentsId"`
	IsWeekStar       bool     `json:"-"`
	IsHoroscopeStar  bool     `json:"-"`
	IsMonthStar      bool     `json:"-"`
	IsModelStudent   bool     `json:"-"`
}

type LiveCache struct {
	Live           LiveProfile `json:"live"`
	ScoreStr       string      `json:"score"`
	Score          float32     `json:"-"`
	FansCount      int         `json:"fansCount"`
	Priority       float32     `json:"priority"`
	Recommand      int         `json:"recommand"`
	RecommandLevel int         `json:"recommandLevel"`
	StarsCount     int         `json:"starsCount"`
	TopCount       int         `json:"topCount"`
	BottomScore    int         `json:"bottomScore"`
	NowIncoming    float32     `json:"nowGem"`
	Lat            float32     `json:"lat"`
	Lng            float32     `json:"lng"`
	IsShowAdd      int         `json:"is_show_add"`
	DayIncoming    float32     `json:"dayIncoming"`
	MonthIncoming  float32     `json:"monthIncoming"`
	Data4Api       interface{} `json:"data"` // 20200305专门为api接口新增的透传参数
	IsGroupVideo   bool        `json:"-"`
}

func (self *LiveCache) GetBusinessScore() float32 {
	var score float32 = 0
	score += self.scoreFx(self.DayIncoming) * 0.2
	score += self.scoreFx(self.MonthIncoming) * 0.05
	score += self.scoreFx(self.Score) * 0.55
	score += self.scoreFx(float32(self.FansCount)) * 0.10
	score += self.scoreFx(float32(self.Live.ShareCount)) * 0.10
	return score
}

func (self *LiveCache) scoreFx(score float32) float32 {
	return (score / 200) / (1 + score/200)
}

func (self *LiveCache) CheckDataType() {
	self.Score = float32(utils.GetFloat64(self.ScoreStr))
	self.Live.UserId = utils.GetInt64(self.Live.UserIdStr)
	self.Live.GemProfit = float32(utils.GetFloat64(self.Live.GemProfitStr))
}

type LiveCacheModule struct {
	cacheLive cache.Cache
}

// 根据liveids 获取直播间信息，如果liveids为空 返回所有直播间
func (lcm *LiveCacheModule) QueryByLiveIds(liveIds []int64) ([]LiveCache, error) {
	lives := make([]LiveCache, 0)
	allList, err := lcm.QueryLiveList()
	if err == nil {
		lives = lcm.MgetByLiveIds(allList, liveIds)
	}
	return lives, err
}

// 根据liveids 获取直播间信息，如果liveids为空 返回所有直播间
func (lcm *LiveCacheModule) MgetByLiveIds(allList []LiveCache, liveIds []int64) []LiveCache {
	live_ids_map := make(map[int64]int)
	if liveIds != nil && len(liveIds) > 0 {
		for _, liveId := range liveIds {
			if liveId > 0 {
				live_ids_map[liveId] = 1
			}
		}
	}

	lives := make([]LiveCache, 0)
	if allList != nil && len(allList) > 0 {
		if len(live_ids_map) == 0 {
			lives = allList
		} else {
			for i := range allList {
				if _, ok := live_ids_map[allList[i].Live.UserId]; ok {
					lives = append(lives, allList[i])
				}
			}
		}
	}
	return lives
}

// 获取所有直播列表
func (lcm *LiveCacheModule) QueryLiveList() ([]LiveCache, error) {
	var initialTime = time.Now()
	var startTime = time.Now()
	list_key := "{cluster1}hotlives_with_recommend_v2"
	live_bytes, err := lcm.cacheLive.LRange(list_key, 0, -1)
	var startJsonTime = time.Now()
	lives := make([]LiveCache, 0)
	for i := 0; i < len(live_bytes); i++ {
		live_byte := live_bytes[i]
		if live_byte != nil && len(live_byte) > 0 {
			live := LiveCache{}
			if err := json.Unmarshal(live_byte, &live); err != nil {
				log.Error(err.Error(), string(live_byte))
			} else {
				live.CheckDataType()
				if live.Live.UserId > 0 {
					lives = append(lives, live)
				}
			}
		}
	}
	var endTime = time.Now()
	log.Debugf("QueryLiveList,all:%.3f,len:%d,cache:%.3f,json:%.3f,week_star:%.3f",
		endTime.Sub(startTime).Seconds(), len(lives),
		startJsonTime.Sub(startTime).Seconds(),
		endTime.Sub(startJsonTime).Seconds(),
		startTime.Sub(initialTime).Seconds())
	return lives, err
}

func (lcm *LiveCacheModule) GetWeekStars() ([]int64, error) {
	key := "live_week_star_recommend"
	var value string
	var users []int64
	err := help.GetStructByCache(lcm.cacheLive, key, &value)
	if err == nil {
		value = strings.Trim(value, "\"")
		for _, single := range strings.Split(value, ",") {
			uid64 := utils.GetInt64(single)
			if uid64 > 0 {
				users = append(users, uid64)
			}
		}
		//log.Debugf("get model student: %+v", users)
		return users, nil
	}
	return users, err
}

func (lcm *LiveCacheModule) GetHoroscopeStars() ([]int64, error) {
	key := "live_horoscope_star_recommend"
	var value string
	var users []int64
	err := help.GetStructByCache(lcm.cacheLive, key, &value)
	if err == nil {
		value = strings.Trim(value, "\"")
		for _, single := range strings.Split(value, ",") {
			uid64 := utils.GetInt64(single)
			if uid64 > 0 {
				users = append(users, uid64)
			}
		}
		//log.Debugf("get model student: %+v", users)
		return users, nil
	}
	return users, err
}

func (lcm *LiveCacheModule) GetMonthStar() (int64, error) {
	key := "live_month_star_recommend"
	var uid int64
	err := help.GetStructByCache(lcm.cacheLive, key, &uid)
	return uid, err
}

func (lcm *LiveCacheModule) GetModelStudents() ([]int64, error) {
	key := "live_icon_model_student"
	var value string
	var users []int64
	err := help.GetStructByCache(lcm.cacheLive, key, &value)
	if err == nil {
		value = strings.Trim(value, "\"")
		for _, single := range strings.Split(value, ",") {
			uid64 := utils.GetInt64(single)
			if uid64 > 0 {
				users = append(users, uid64)
			}
		}
		//log.Debugf("get model student: %+v", users)
		return users, nil
	}
	return users, err
}

func NewLiveCacheModule(cache *cache.Cache) *LiveCacheModule {
	return &LiveCacheModule{cacheLive: *cache}
}

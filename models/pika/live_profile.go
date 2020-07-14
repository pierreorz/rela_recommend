package pika

import (
	"encoding/json"
	"rela_recommend/cache"
	"rela_recommend/log"
	"rela_recommend/utils"
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
	DayIncoming    float32     `json:"dayIncoming"`
	MonthIncoming  float32     `json:"monthIncoming"`
	Data4Api       interface{} `json:"data"` // 20200305专门为api接口新增的透传参数
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
func (self *LiveCacheModule) QueryByLiveIds(liveIds []int64) ([]LiveCache, error) {
	lives := make([]LiveCache, 0)
	allList, err := self.QueryLiveList()
	if err == nil {
		lives = self.MgetByLiveIds(allList, liveIds)
	}
	return lives, err
}

// 根据liveids 获取直播间信息，如果liveids为空 返回所有直播间
func (self *LiveCacheModule) MgetByLiveIds(allList []LiveCache, liveIds []int64) []LiveCache {
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
			for i, _ := range allList {
				if _, ok := live_ids_map[allList[i].Live.UserId]; ok {
					lives = append(lives, allList[i])
				}
			}
		}
	}
	return lives
}

// 获取所有直播列表
func (self *LiveCacheModule) QueryLiveList() ([]LiveCache, error) {
	var startTime = time.Now()
	list_key := "{cluster1}hotlives_with_recommend_v2"
	live_bytes, err := self.cacheLive.LRange(list_key, 0, -1)
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
	log.Infof("QueryLiveList,all:%.3f,len:%d,cache:%.3f,json:%.3f",
		endTime.Sub(startTime).Seconds(), len(lives),
		startJsonTime.Sub(startTime).Seconds(),
		endTime.Sub(startJsonTime).Seconds())
	return lives, err
}

func NewLiveCacheModule(cache *cache.Cache) *LiveCacheModule {
	return &LiveCacheModule{cacheLive: *cache}
}

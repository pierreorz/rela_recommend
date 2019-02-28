package pika

import (
	"encoding/json"
	"rela_recommend/log"
	"rela_recommend/cache"
)

type LiveProfile struct {
	LiveId         		int64		`json:"id"`         // 用户ID
	LiveTypeId     		int 		`json:"liveTypeId"`            
	UserId				int64		`json:"userId"`
	Text				string		`json:"text"`
	CreateTime			JsonTime	`json:"createTime"`
	UpdateTime			JsonTime	`json:"updateTime"`
	Active				int			`json:"active"`
	TopView				int			`json:"topView"`
	Ip					string		`json:"ip"`
	Ua					string		`json:"ua"`
	Lat					float32		`json:"lat"`
	Lng					float32		`json:"lng"`
	City				string		`json:"city"`
	GemProfit			float32		`json:"gemProfit"`
	SendMsgCount		int			`json:"sendMsgCount"`
	ReceivedMsgCount 	int			`json:"receivedMsgCount"`
	ShareCount			int			`json:"shareCount"`
	AudioType			int			`json:"audioType"`
	IsMulti				int			`json:"isMulti"`
}

type LiveCache struct {
	Live 				LiveProfile	`json:"live"`
	Score				float32		`json:"score"`
	FansCount			int			`json:"fansCount"`
	Priority			float32		`json:"priority"`
	DayIncoming			float32		`json:"dayIncoming"`
	MonthIncoming		float32		`json:"monthIncoming"`
}


type LiveCacheModule struct {
	cacheLive cache.Cache
}

// 根据liveids 获取直播间信息，如果liveids为空 返回所有直播间
func (self *LiveCacheModule) QueryByLiveIds(liveIds []int64) ([]LiveCache, error) {
	live_ids_map := make(map[int64]int)
	if liveIds != nil && len(liveIds) > 0 {
		for _, liveId := range liveIds {
			live_ids_map[liveId] = 1
		}
	}

	list_key := "hotlives_with_recommend_v2"
	live_strs, err := self.cacheLive.LRange(list_key, 0, -1)
	lives := make([]LiveCache, 0)
	for i := 0; i < len(live_strs); i++ {
		live_str := live_strs[i]
		if len(live_str) > 0 {
			live := LiveCache{}
			if err := json.Unmarshal(([]byte)(live_str), &live); err != nil {
				log.Error(err.Error(), live_str)
			} else if live.Live.UserId > 0	{
				if len(live_ids_map) == 0 {
					lives = append(lives, live)
				} else if _, ok := live_ids_map[live.Live.UserId]; ok {
					lives = append(lives, live)
				}
			}
		}
	}
	return lives, err
}


func NewLiveCacheModule(cache *cache.Cache) *LiveCacheModule {
	return &LiveCacheModule{cacheLive: *cache}
}



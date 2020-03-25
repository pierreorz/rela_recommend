package live

import (
	"fmt"
	"time"
	"sync"
	"errors"
	"rela_recommend/factory"
	"rela_recommend/log"
	"rela_recommend/rpc/api"
	"rela_recommend/models/pika"
)


var cachedLiveListMap map[int][]pika.LiveCache = map[int][]pika.LiveCache{}
var cachedLiveListMapUpdateLocker = &sync.RWMutex{}

// 获取缓存的直播列表
func GetCachedLiveMapList(liveType int) []pika.LiveCache {
	return cachedLiveListMap[liveType]
}

// 获取缓存的直播用户id为key 的 map
func GetCachedLiveMapMap(liveType int) map[int64]*pika.LiveCache {
	liveMap := map[int64]*pika.LiveCache{}
	liveList := GetCachedLiveMapList(liveType)
	for i, _ := range liveList {
		liveMap[liveList[i].Live.UserId] = &liveList[i]
	}
	return liveMap
}

func convertApiLive2RedisLiveList(lives []api.SimpleChatroom) []pika.LiveCache {
	liveCacheList := make([]pika.LiveCache, len(lives))
	for i, live := range lives {
		liveCache := &liveCacheList[i]

		liveCache.Live.LiveId         		= live.UserID
		liveCache.Live.LiveTypeId     		= live.LiveType
		// liveCache.Live.UserIdStr			= live.UserId
		liveCache.Live.UserId				= live.UserID
		// liveCache.Live.Text					= live.
		liveCache.Live.CreateTime			= pika.JsonTime{ Time: live.CreateTime }
		// liveCache.Live.UpdateTime			= live.
		// liveCache.Live.Active				= live.
		liveCache.Live.TopView				= live.TopView
		// liveCache.Live.Ip					= live.
		// liveCache.Live.Ua					= live.
		liveCache.Live.Lat					= live.Lat
		liveCache.Live.Lng					= live.Lng
		// liveCache.Live.City					= live.
		// liveCache.Live.GemProfitStr			= live.
		liveCache.Live.GemProfit			= live.GemProfit
		liveCache.Live.SendMsgCount			= live.SendMsgCount
		liveCache.Live.ReceivedMsgCount 	= live.ReceivedMsgCount
		liveCache.Live.ShareCount			= live.ShareCount
		// liveCache.Live.AudioType			= live.
		liveCache.Live.IsMulti				= live.IsMulti

		// liveCache.ScoreStr			= live.Score
		liveCache.Score				= live.Score
		liveCache.FansCount			= live.FansCount
		liveCache.Priority			= live.Priority
		liveCache.Recommand			= live.Recommend
		liveCache.RecommandLevel	= live.RecommendLevel
		liveCache.StarsCount		= live.StarsCount
		liveCache.TopCount			= live.TopCount
		liveCache.BottomScore		= live.BottomScore
		liveCache.DayIncoming		= live.DayIncoming
		liveCache.MonthIncoming		= live.MonthIncoming
		liveCache.Data4Api			= live.Data
	}
	return liveCacheList
}

func updateCachedLiveMap(liveType int, newList []api.SimpleChatroom) (int, int, error) {
	var err error
	oldList := GetCachedLiveMapList(liveType)
	oldLen := len(oldList)
	newLen := len(newList)

	// 降级保障策略：原列表大于20个但新列表突然为空，认定为错误列表，不进行更新
	if oldLen >= 20 && newLen == 0 {
		errMsg := fmt.Sprintf("refreshLiveList api err type %d: old %d but new is empty!\n", liveType, len(oldList))
		err = errors.New(errMsg)
		log.Warnf(errMsg)
	} else {
		newCacheList := convertApiLive2RedisLiveList(newList)

		cachedLiveListMapUpdateLocker.Lock()
		defer cachedLiveListMapUpdateLocker.Unlock()
	
		cachedLiveListMap[liveType] = newCacheList
	}
	return oldLen, newLen, err
}

func refreshLiveMapList(duration time.Duration) {
	// time.Sleep(10 * time.Second)
	log.Infof("refreshLiveList api task start: %s\n", duration)
	tick := time.NewTicker(duration)
	for {
		select {
		case <- tick.C:
			if factory.ChatRoomRpcClient != nil {
				var startTime = time.Now()
				var updateLog []string = []string{}
				for _, liveType := range api.ChatRoomLiveTypes {
					var typeTimeStart = time.Now()
					lives, err := api.CallChatRoomList(liveType)
					var typeTimeTotal = time.Now().Sub(typeTimeStart).Seconds()
					if err == nil {
						oldLen, newLen, err2 := updateCachedLiveMap(liveType, lives)

						updateLog = append(updateLog, fmt.Sprintf("type %d time %.3f old %d new %d err %s", liveType, typeTimeTotal, oldLen, newLen, err2))
					} else {
						updateLog = append(updateLog, fmt.Sprintf("type %d time %.3f err %s", liveType, typeTimeTotal, err))
					}
				}
				var endTime = time.Now()
				log.Infof("refreshLiveList api over %.3f: %+v\n", endTime.Sub(startTime).Seconds(), updateLog)
			} else {
				log.Warnf("refreshLiveList err:api is not ready\n")
			}
		}
	}
}

// 启动自动刷新直播列表缓存
func init() {
	go refreshLiveMapList(2 * time.Second)
}
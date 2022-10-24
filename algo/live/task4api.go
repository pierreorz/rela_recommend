package live

import (
	"encoding/json"
	"errors"
	"fmt"
	"rela_recommend/factory"
	"rela_recommend/log"
	"rela_recommend/models/pika"
	"rela_recommend/rpc/api"
	"rela_recommend/service/performs"
	"rela_recommend/utils"
	"sync"
	"time"
)

var cachedLiveListMap map[int][]pika.LiveCache = map[int][]pika.LiveCache{}
var cachedLiveListMapUpdateLocker = &sync.RWMutex{}

// 获取缓存的直播列表 -1.all; 1. video; 2. audio; 3. multi_audio(radio)
// https://wiki.rela.me/pages/viewpage.action?pageId=5672757
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

// 通过直播分类获取直播列表
func GetCachedLiveListByTypeClassify(typeId int, classify int) []pika.LiveCache {
	// 客户端和服务端约定了两个大数分别代表 video 和 multi_audio
	if classify == typeBigVideo {
		typeId = 1
		classify = typeRecommend
	} else if classify == typeBigMultiAudio {
		typeId = 3
		classify = typeRecommend
	}
	var lives []pika.LiveCache
	if typeId <= 0 {
		typeId = -1
	}
	liveList := GetCachedLiveMapList(typeId) // -1 获取所有直播
	for i, live := range liveList {
		if classify == typeGroupVideo {
			// 二十大会特殊审核逻辑，把叽叽喳喳单独列出来
			if live.IsGroupVideo {
				lives = append(lives, liveList[i])
			}
		} else if classify <= typeRecommend || live.Live.Classify == classify {
			// 不限制 或 类型是特定类型; classify=1时为推荐分类，不限制分类
			lives = append(lives, liveList[i])
		}
	}
	return lives
}

// 通过直播分类获取直播开播日志列表
func GetCachedLiveMomentListByTypeClassify(typeId int, classify int) map[int64]int {
	lives := GetCachedLiveListByTypeClassify(typeId, classify)
	MomScoreMap := make(map[int64]float64, 0)
	for _, live := range lives {
		MomScoreMap[live.Live.MomentsID] = float64(live.GetBusinessScore())
	}
	MomRankMap := make(map[int64]int, 0)
	Moms := utils.SortMapByValue(MomScoreMap)
	for rank, id := range Moms {
		MomRankMap[id] = rank + 1
	}
	return MomRankMap
}

func convertApiLive2RedisLiveList(lives []api.SimpleChatroom) []pika.LiveCache {
	liveCacheClient := pika.NewLiveCacheModule(&factory.CacheLiveRds)
	weekStars, err := liveCacheClient.GetWeekStars()
	if err != nil {
		log.Errorf("get week star error: %s", err)
	}

	monthStarUID, err := liveCacheClient.GetMonthStar()
	if err != nil {
		log.Errorf("get month star error: %s", err)
	}

	modelStudents, err := liveCacheClient.GetModelStudents()
	if err != nil {
		log.Errorf("get model student error: %s", err)
	}

	horoscopeStars, err := liveCacheClient.GetHoroscopeStars()
	if err != nil {
		log.Errorf("get horoscope star error: %s", err)
	}

	liveCacheList := make([]pika.LiveCache, len(lives))
	for i, live := range lives {
		liveCache := &liveCacheList[i]

		liveCache.Live.LiveId = live.UserID
		liveCache.Live.LiveTypeId = live.LiveType
		// liveCache.Live.UserIdStr			= live.UserId
		liveCache.Live.UserId = live.UserID
		// liveCache.Live.Text					= live.
		liveCache.Live.CreateTime = pika.JsonTime{Time: live.CreateTime}
		// liveCache.Live.UpdateTime			= live.
		// liveCache.Live.Active				= live.
		liveCache.Live.TopView = live.TopView
		// liveCache.Live.Ip					= live.
		// liveCache.Live.Ua					= live.
		liveCache.Live.Lat = live.Lat
		liveCache.Live.Lng = live.Lng
		// liveCache.Live.City					= live.
		// liveCache.Live.GemProfitStr			= live.
		liveCache.Live.GemProfit = live.GemProfit
		liveCache.Live.SendMsgCount = live.SendMsgCount
		liveCache.Live.ReceivedMsgCount = live.ReceivedMsgCount
		liveCache.Live.ShareCount = live.ShareCount
		// liveCache.Live.AudioType			= live.
		liveCache.Live.IsMulti = live.IsMulti
		liveCache.Live.Classify = live.Classify
		liveCache.Live.MomentsID = live.MomentsID
		if len(weekStars) > 0 {
			contained := utils.ContainsInt64(weekStars, liveCache.Live.UserId)
			if contained {
				liveCache.Live.IsWeekStar = true
			}
		}

		if len(horoscopeStars) > 0 {
			contained := utils.ContainsInt64(horoscopeStars, liveCache.Live.UserId)
			if contained {
				liveCache.Live.IsHoroscopeStar = true
			}
		}
		if liveCache.Live.UserId == monthStarUID {
			liveCache.Live.IsMonthStar = true
		}
		if len(modelStudents) >= 0 {
			contained := utils.ContainsInt64(modelStudents, liveCache.Live.UserId)
			if contained {
				liveCache.Live.IsModelStudent = true
			}
		}

		// liveCache.ScoreStr			= live.Score
		liveCache.Score = live.Score
		liveCache.FansCount = live.FansCount
		liveCache.Priority = live.Priority
		liveCache.Recommand = live.Recommend
		liveCache.RecommandLevel = live.RecommendLevel
		liveCache.StarsCount = live.StarsCount
		liveCache.TopCount = live.TopCount
		liveCache.BottomScore = live.BottomScore
		liveCache.NowIncoming = live.NowIncoming
		liveCache.Lat = live.Lat
		liveCache.Lng = live.Lng
		liveCache.IsShowAdd = live.IsShowAdd
		liveCache.DayIncoming = live.DayIncoming
		liveCache.MonthIncoming = live.MonthIncoming
		liveCache.Data4Api = live.Data

		var item ILiveRankItemV3
		err := json.Unmarshal([]byte(live.Data), &item)
		if err == nil {
			switch item.Status {
			case MultiVideoFour, MultiVideoNine:
				liveCache.IsGroupVideo = true
			}
		}
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
		pf := &performs.Performs{}
		select {
		case <-tick.C:
			pf.Run("rpcs", func(perform *performs.Performs) interface{} {
				if factory.ChatRoomRpcClient != nil {
					for _, liveType := range api.ChatRoomLiveTypes {
						pf.Run(fmt.Sprintf("type%d", liveType), func(perform *performs.Performs) interface{} {
							lives, err := api.CallChatRoomList(liveType)
							if err == nil {
								if _, newLen, err2 := updateCachedLiveMap(liveType, lives); err2 == nil {
									return newLen
								} else {
									return err2
								}
							} else {
								return err
							}
						})
					}
				} else {
					return errors.New("api not ready")
				}
				return nil
			})
		}

		log.Debugf("algo.task:live.rpc:%s\n", pf.ToString())
		pf.ToWriteChan("algo.task", map[string]string{
			"app": "live.rpc",
		}, map[string]interface{}{}, *pf.BeginTime)
	}
}

// 启动自动刷新直播列表缓存
func init() {
	go refreshLiveMapList(5 * time.Second)
}

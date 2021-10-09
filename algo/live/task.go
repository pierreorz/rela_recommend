package live

import (
	"encoding/json"
	"errors"
	"fmt"
	"rela_recommend/factory"
	"rela_recommend/help"
	"rela_recommend/log"
	"rela_recommend/models/pika"
	"rela_recommend/service/performs"
	"time"
)

var cachedLiveList []pika.LiveCache = make([]pika.LiveCache, 0)

// 获取缓存的直播列表
func GetCachedLiveList() []pika.LiveCache {
	return cachedLiveList
}

// 获取缓存的直播用户id为key 的 map
func GetCachedLiveMap() map[int64]*pika.LiveCache {
	liveMap := map[int64]*pika.LiveCache{}
	liveList := GetCachedLiveList() // 调用方法获取对象copy，避免引发bug
	for i, _ := range liveList {
		liveMap[liveList[i].Live.UserId] = &liveList[i]
	}
	return liveMap
}

func refreshLiveList(duration time.Duration) {
	// time.Sleep(10 * time.Second)
	log.Infof("refreshLiveList task start: %s\n", duration)
	tick := time.NewTicker(duration)
	for {
		pf := &performs.Performs{}
		select {
		case <-tick.C:
			pf.Run("cache", func(perform *performs.Performs) interface{} {
				if factory.CacheLiveRds != nil {
					liveCache := pika.NewLiveCacheModule(&factory.CacheLiveRds)
					lives, err := liveCache.QueryLiveList()
					if err != nil {
						return err
					} else {
						if lives != nil {
							// 防止缓存临时挂掉引起列表为空： 原列表>=20时，新列表突然为0时不更新，假如有脏数据，外层生成列表时会校验。
							if len(cachedLiveList) >= 20 && len(lives) == 0 {
								return errors.New(fmt.Sprintf("old %d, new %d", len(cachedLiveList), len(lives)))
							} else {
								cachedLiveList = lives
								return len(cachedLiveList)
							}
						} else {
							return errors.New(fmt.Sprintf("list is nil"))
						}
					}
				} else {
					return errors.New("cache not ready")
				}
			})

			pf.Run("cache_classify", func(perform *performs.Performs) interface{} {
				if factory.CacheInternalRds != nil {
					var classifyData []ClassifyItem
					err := help.GetStructByCache(factory.CacheInternalRds, "liveClassify:data:", &classifyData)
					if err == nil {
						tempMap := make(map[int]multiLanguage, len(classifyData))
						for _, item := range classifyData {
							var multi multiLanguage
							err = json.Unmarshal([]byte(item.SkillsJson), &multi)
							if err == nil {
								tempMap[int(item.Id)] = multi
							} else {
								log.Warnf("live classify unmarshal %s err: %+v", item.SkillsJson, err)
							}
						}
						classifyMap = tempMap
						return len(classifyMap)
					} else {
						return err
					}
				} else {
					return errors.New("cache not ready")
				}
			})
		}

		log.Debugf("algo.task:live.cache:%s\n", pf.ToString())
		pf.ToWriteChan("algo.task", map[string]string{
			"app": "live.cache",
		}, map[string]interface{}{}, *pf.BeginTime)
	}
}

// 启动自动刷新直播列表缓存
func init() {
	go refreshLiveList(2 * time.Second)
}

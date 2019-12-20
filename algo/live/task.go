package live

import (
	"time"
	"rela_recommend/log"
	"rela_recommend/factory"
	"rela_recommend/models/pika"
)

var cachedLiveList []pika.LiveCache = make([]pika.LiveCache, 0)

// 获取缓存的直播列表
func GetCachedLiveList() []pika.LiveCache {
	return cachedLiveList
}

// 获取缓存的直播用户id为key 的 map
func GetCachedLiveMap() map[int64]*pika.LiveCache {
	liveMap := map[int64]*pika.LiveCache{}
	liveList := GetCachedLiveList()  // 调用方法获取对象copy，避免引发bug
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
		select {
		case <- tick.C:
			if factory.CacheLiveRds != nil {
				var startTime = time.Now()
				liveCache := pika.NewLiveCacheModule(&factory.CacheLiveRds)
				lives, err := liveCache.QueryLiveList()
				var endTime = time.Now()
				if err != nil {
					log.Errorf("refreshLiveList err %.3f: %s\n", endTime.Sub(startTime).Seconds(), err)
				} else {
					if lives != nil {
						// 防止缓存临时挂掉引起列表为空： 原列表>=20时，新列表突然为0时不更新，假如有脏数据，外层生成列表时会校验。
						if len(cachedLiveList) >= 20 && len(lives) == 0 {
							log.Infof("refreshLiveList err %.3f: old %d, new %d\n", endTime.Sub(startTime).Seconds(), len(cachedLiveList), len(lives))
						} else {
							cachedLiveList = lives
							log.Infof("refreshLiveList over %.3f: %d\n", endTime.Sub(startTime).Seconds(), len(cachedLiveList))
						}
					} else {
						log.Errorf("refreshLiveList err %.3f: list is nil\n", endTime.Sub(startTime).Seconds())
					}
				}
			} else {
				log.Errorf("refreshLiveList err:cache is not ready\n")
			}
		}
	}
}

// 启动自动刷新直播列表缓存
func init() {
	go refreshLiveList(2 * time.Second)
}
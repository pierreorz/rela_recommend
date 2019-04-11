package live

import (
	"time"
	"rela_recommend/log"
	"rela_recommend/factory"
	"rela_recommend/models/pika"
)

var CachedLiveList []pika.LiveCache = make([]pika.LiveCache, 0)

func refreshLiveList(duration time.Duration) {
	log.Infof("refreshLiveList task start: %s\n", duration)
	tick := time.NewTicker(duration)
	liveCache := pika.NewLiveCacheModule(&factory.CacheLiveRds)
	for {
		select {
		case <- tick.C:
			var startTime = time.Now()
			lives, err := liveCache.QueryLiveList()
			var endTime = time.Now()
			if err != nil {
				log.Errorf("refreshLiveList err %.3f: %s\n", endTime.Sub(startTime).Seconds(), err)
			} else {
				CachedLiveList = lives
				log.Infof("refreshLiveList over %.3f: %d\n", endTime.Sub(startTime).Seconds(), len(CachedLiveList))
			}
		}
	}
}

// 启动自动刷新直播列表缓存
func init() {
	time.Sleep(10 * time.Second)
	liveCache := pika.NewLiveCacheModule(&factory.CacheLiveRds)
	lives, err := liveCache.QueryLiveList()
	log.Infof("live init task start %d: %s\n", len(lives), err)
	go refreshLiveList(2 * time.Second)
}
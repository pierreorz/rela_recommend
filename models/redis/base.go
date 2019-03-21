package redis

import (
	"fmt"
	"time"
	"rela_recommend/log"
	"rela_recommend/cache"
)


type CachePikaModule struct {
	cache cache.Cache
	store cache.Cache
}

func (self *CachePikaModule) Get(key string) (interface{}, error) {
	res, err := self.cache.Get(key)
	if err != nil {
		log.Warnf("CachePikaModule  get warn: %s\n", err)
	}
	if res == nil {
		res, err = self.store.Get(key)
	}
	return res, err
}

func (self *CachePikaModule) MGetSet(ids []int64, keyFormater string, cacheTime int64)([]interface{}, error) {
	var startTime = time.Now()
	dataLen := len(ids)
	// 构造keys
	keys := make([]string, dataLen)
	for i, id := range ids {
		keys[i] = fmt.Sprintf(keyFormater, id)
	}
	var startCacheTime = time.Now()
	// 从缓存读取
	ress, err := self.cache.Mget(keys)
	var endCacheTime = time.Now()
	if err != nil {
		log.Warnf("CachePikaModule mget warn: %s\n", err)
	}
	var notFoundIndexs = make([]int, 0)
	var notFoundKeys = make([]string, 0)
	for i, res := range ress {
		if res == nil {
			notFoundIndexs = append(notFoundIndexs, i)
			notFoundKeys = append(notFoundKeys, keys[i])
		}
	}
	var startStoreTime = time.Now()
	var startStoreSetTime = time.Now()
	if len(notFoundIndexs) > 0 {
		// 从持久化存储读取
		ress2, err2 := self.store.Mget(notFoundKeys)
		if err2 == nil {
			var setKeyVals = make(map[string]interface{}, 0)
			for i, res := range ress2 {
				if res != nil {
					ress[notFoundIndexs[i]] = res
					setKeyVals[notFoundKeys[i]] = res
				}
			}
			if len(setKeyVals) > 0 {  // 设置缓存
				startStoreSetTime = time.Now()
				self.cache.MsetEx(setKeyVals, cacheTime)
			}
		} else {
			err = err2
		}
	}
	var endTime = time.Now()
	log.Infof("ReadKey:%s,all:%d,cache:%d,store:%d,final:%d;total:%.3f,keys:%.3f,cache:%.3f,notfound:%.3f,store:%.3f,2cache:%.3f\n",
		keyFormater, dataLen, dataLen-len(notFoundIndexs), len(notFoundIndexs),len(ress), 
		endTime.Sub(startTime).Seconds(),
		startCacheTime.Sub(startTime).Seconds(), endCacheTime.Sub(startCacheTime).Seconds(),
		startStoreTime.Sub(endCacheTime).Seconds(), startStoreSetTime.Sub(startStoreTime).Seconds(),
		endTime.Sub(startStoreSetTime).Seconds())
	return ress, err
}
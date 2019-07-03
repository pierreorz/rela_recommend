package redis

import (
	"fmt"
	"sync"
	"time"
	"reflect"
	"encoding/json"
	"rela_recommend/log"
	"rela_recommend/cache"
	"rela_recommend/utils"
	"rela_recommend/algo"
)


type CachePikaModule struct {
	cache cache.Cache
	store cache.Cache
	ctx algo.IContext
}

// 读取缓存。cacheTime：redis的缓存时间；cacheNilTime: pika也不存在的key写入redis的缓存时间
func (self *CachePikaModule) GetSet(id int64, keyFormater string, cacheTime int, cacheNilTime int) (interface{}, error) {
	var startTime = time.Now()
	var dataFrom = "cache"
	var message = ""
	key := fmt.Sprintf(keyFormater, id)
	res, err := self.cache.Get(key)
	var endCacheTime = time.Now()
	if err != nil {
		message = fmt.Sprintf("cache warn: %s\n", err)
	}
	var startStoreTime = time.Now()
	var startStoreSetTime = time.Now()
	if res == nil {  // 缓存没有获取到
		res, err = self.store.Get(key)
		startStoreSetTime = time.Now()
		if err == nil {
			dataFrom = "store"
			res2, cacheTime2 := res, cacheTime
			if res == nil {
				res2, cacheTime2 = "", cacheNilTime
			}
			err = self.cache.SetEx(key, res2, cacheTime2)
			if err != nil {
				message = fmt.Sprintf("%s;setcache warn: %s\n", message, err)
			}
		} else {
			message = fmt.Sprintf("%s;store warn: %s\n", message, err)
		}
	}
	var endTime = time.Now()
	log.Infof("GetSet Key:%s,id:%d,from:%s,total:%.3f,cache:%.3f,store:%.3f,2cache:%.3f,msg:%s\n",
		keyFormater, id, dataFrom,
		endTime.Sub(startTime).Seconds(), endCacheTime.Sub(startTime).Seconds(), 
		startStoreSetTime.Sub(startStoreTime).Seconds(), endTime.Sub(startStoreSetTime).Seconds(),
		message)
	return res, err
}

// 读取缓存。cacheTime：redis的缓存时间；cacheNilTime: pika也不存在的key写入redis的缓存时间
func (self *CachePikaModule) MGetSet(ids []int64, keyFormater string, cacheTime int64, cacheNilTime int64)([]interface{}, error) {
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
		} else {
			// 处理空字符串
			if bs, ok := res.([]byte); ok && len(bs) == 0 {
				ress[i] = nil
			}
		}
	}
	var startStoreTime = time.Now()
	var startStoreSetTime = time.Now()
	var setLen, setNilLen int = 0, 0
	if len(notFoundIndexs) > 0 {
		// 从持久化存储读取
		ress2, err2 := self.store.Mget(notFoundKeys)
		if err2 == nil {
			var setKeyVals = make(map[string]interface{}, 0)
			var setNilVals = make(map[string]interface{}, 0)
			for i, res := range ress2 {
				if res != nil {
					ress[notFoundIndexs[i]] = res
					setKeyVals[notFoundKeys[i]] = res
				} else {
					setNilVals[notFoundKeys[i]] = ""
				}
			}
			setLen, setNilLen = len(setKeyVals), len(setNilVals)
			startStoreSetTime = time.Now()
			if cacheTime > 0 && setLen > 0 {  // 设置缓存
				self.cache.MsetEx(setKeyVals, cacheTime)
			}
			if cacheNilTime > 0 && setNilLen > 0 {  // 设置找不到的缓存
				self.cache.MsetEx(setNilVals, cacheNilTime)
			}
		} else {
			err = err2
		}
	}
	var endTime = time.Now()
	log.Infof("ReadKey:%s,all:%d,cache:%d,store:%d,final:%d,set:%d,setnil:%d;total:%.3f,keys:%.3f,cache:%.3f,notfound:%.3f,store:%.3f,2cache:%.3f\n",
		keyFormater, dataLen, dataLen-len(notFoundIndexs), len(notFoundIndexs),len(ress), setLen, setNilLen,
		endTime.Sub(startTime).Seconds(),
		startCacheTime.Sub(startTime).Seconds(), endCacheTime.Sub(startCacheTime).Seconds(),
		startStoreTime.Sub(endCacheTime).Seconds(), startStoreSetTime.Sub(startStoreTime).Seconds(),
		endTime.Sub(startStoreSetTime).Seconds())
	return ress, err
}


func (self *CachePikaModule) jsonsToValues(jsons []interface{}, objType reflect.Type) ([]reflect.Value, error) {
	objs := make([]reflect.Value, 0, len(jsons))
	for _, res := range jsons {
		if res != nil {
			var newObj = reflect.New(objType)
			bs, ok := res.([]byte)
			if ok {
				if err := json.Unmarshal(bs, newObj.Interface()); err == nil {
					objs = append(objs, reflect.Indirect(newObj))
				} else {
					log.Warn("json err:", res , err.Error())
				}
			} else {
				log.Warn("must []byte:", res)
			}
		}
	}
	return objs, nil
}

func (self *CachePikaModule) Jsons2StructsBySingle(jsons []interface{}, obj interface{}) (*reflect.Value, error) {
	startTime := time.Now()
	objLen := len(jsons)
	objTyp := reflect.TypeOf(obj)
	objSlc := reflect.MakeSlice(reflect.SliceOf(objTyp), 0, objLen)
	ress, err := self.jsonsToValues(jsons, objTyp)
	for _, res := range ress {
		objSlc = reflect.Append(objSlc, res)
	}
	endTime := time.Now()
	log.Infof("Jsons2StructsBySingle all:%d,notfound:%d,final:%d;total:%.4f\n",
		len(jsons), len(jsons)-objSlc.Len(), objSlc.Len(), 
		endTime.Sub(startTime).Seconds())
	return &objSlc, err
}

func (self *CachePikaModule) Jsons2StructsByRoutine(jsons []interface{}, obj interface{}, partLen int) (*reflect.Value, error) {
	startTime := time.Now()

	parts := utils.SplitList(jsons, partLen)
	objTyp := reflect.TypeOf(obj)
	objSlc := reflect.MakeSlice(reflect.SliceOf(objTyp), 0, len(jsons))
	var err error

	var lockHandler = &sync.RWMutex{}
	wg := new(sync.WaitGroup)
	for _, part := range parts {
		wg.Add(1)
		go func(part []interface{}) {
			defer wg.Done()
			partRes, partErr := self.jsonsToValues(part, objTyp)
			lockHandler.Lock()
			defer lockHandler.Unlock()
			if err == nil {
				err = partErr
			}
			for _, res := range partRes {
				objSlc = reflect.Append(objSlc, res)
			}
		}(part)
	}
	wg.Wait()

	endTime := time.Now()
	log.Infof("Jsons2StructsByRoutine all:%d,notfound:%d,final:%d;total:%.4f\n",
		len(jsons), len(jsons)-objSlc.Len(), objSlc.Len(), 
		endTime.Sub(startTime).Seconds())
	return &objSlc, err
}

func (self *CachePikaModule) Jsons2Structs(jsons []interface{}, obj interface{}) (*reflect.Value, error) {
	abtest := self.ctx.GetAbTest()
	threshold := abtest.GetInt("redis.json.thread.threshold", 200)
	if len(jsons) > threshold {
		jobs := abtest.GetInt("redis.json.thread.jobs", 4)
		return self.Jsons2StructsByRoutine(jsons, obj, jobs)
	} else {
		return self.Jsons2StructsBySingle(jsons, obj)
	}
}

func (self *CachePikaModule) MGetStructs(obj interface{}, ids []int64, keyFormater string, cacheTime int64, cacheNilTime int64) (*reflect.Value, error) {
	startTime := time.Now()
	ress, err := self.MGetSet(ids, keyFormater, cacheTime, cacheNilTime)
	startJsonTime := time.Now()
	objs, err := self.Jsons2Structs(ress, obj)
	endTime := time.Now()
	log.Infof("UnmarshalKey:%s,all:%d,notfound:%d,final:%d;total:%.4f,read:%.4f,json:%.4f\n",
		keyFormater, len(ids), len(ids)-objs.Len(), objs.Len(), 
		endTime.Sub(startTime).Seconds(),
		startJsonTime.Sub(startTime).Seconds(), endTime.Sub(startJsonTime).Seconds())
	return objs, err
}

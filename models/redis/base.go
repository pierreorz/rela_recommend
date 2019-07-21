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
	"rela_recommend/service/abtest"
)


type CachePikaModule struct {
	cache cache.Cache
	store cache.Cache
	ctx algo.IContext
}

// 读取缓存。cacheTime：redis的缓存时间；cacheNilTime: pika也不存在的key写入redis的缓存时间
// 当 cacheTime 和 cacheNilTime 都小雨等于0 则不请求 store
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
	if res == nil && self.store != nil && (cacheTime > 0 || cacheNilTime > 0)  {  // 缓存没有获取到
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
	log.Infof("GetSet rankId:%s,Key:%s,id:%d,from:%s,total:%.3f,cache:%.3f,store:%.3f,2cache:%.3f,msg:%s\n",
		"", keyFormater, id, dataFrom,
		endTime.Sub(startTime).Seconds(), endCacheTime.Sub(startTime).Seconds(), 
		startStoreSetTime.Sub(startStoreTime).Seconds(), endTime.Sub(startStoreSetTime).Seconds(),
		message)
	return res, err
}

func (self *CachePikaModule) Get(id int64, keyFormater string) (interface{}, error) {
	return self.GetSet(id, keyFormater, 0, 0)
}

// 读取缓存。cacheTime：redis的缓存时间，如果<=0则不缓存；cacheNilTime: pika也不存在的key写入redis的缓存时间，如果<=0则不缓存；
// 当 cacheTime 和 cacheNilTime 都小雨等于0 则不请求 store
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
	if len(notFoundIndexs) > 0 && self.store != nil && (cacheTime > 0 || cacheNilTime > 0) {
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
	log.Infof("ReadKey:%s,rankId:%s,all:%d,cache:%d,store:%d,final:%d,set:%d,setnil:%d;total:%.3f,keys:%.3f,cache:%.3f,notfound:%.3f,store:%.3f,2cache:%.3f\n",
		keyFormater, "", dataLen, dataLen-len(notFoundIndexs), len(notFoundIndexs),len(ress), setLen, setNilLen,
		endTime.Sub(startTime).Seconds(),
		startCacheTime.Sub(startTime).Seconds(), endCacheTime.Sub(startCacheTime).Seconds(),
		startStoreTime.Sub(endCacheTime).Seconds(), startStoreSetTime.Sub(startStoreTime).Seconds(),
		endTime.Sub(startStoreSetTime).Seconds())
	return ress, err
}

func (self *CachePikaModule) MGet(ids []int64, keyFormater string)([]interface{}, error) {
	return self.MGetSet(ids, keyFormater, 0, 0)
}

// 将interface类型转化为特定类型，顺序保证
func (self *CachePikaModule) jsonsToValues(jsons []interface{}, objType reflect.Type) ([]*reflect.Value, error) {
	objs := make([]*reflect.Value, len(jsons), len(jsons))
	for i, res := range jsons {
		if res != nil {
			var newObj = reflect.New(objType)
			bs, ok := res.([]byte)
			if ok {
				if err := json.Unmarshal(bs, newObj.Interface()); err == nil {
					newValue := reflect.Indirect(newObj)
					objs[i] = &newValue
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

// 单线程转化json为struct.去除空值，不保证顺序
func (self *CachePikaModule) Jsons2StructsBySingle(jsons []interface{}, obj interface{}) (*reflect.Value, error) {
	startTime := time.Now()
	objLen := len(jsons)
	objTyp := reflect.TypeOf(obj)
	objSlc := reflect.MakeSlice(reflect.SliceOf(objTyp), 0, objLen)
	ress, err := self.jsonsToValues(jsons, objTyp)
	for _, res := range ress {
		if res != nil {
			objSlc = reflect.Append(objSlc, *res)
		}
	}
	endTime := time.Now()
	log.Infof("Jsons2StructsBySingle rankId:%s,all:%d,notfound:%d,final:%d;total:%.4f\n",
		"", len(jsons), len(jsons)-objSlc.Len(), objSlc.Len(), 
		endTime.Sub(startTime).Seconds())
	return &objSlc, err
}

// 多线程转化json为struct.去除空值，不保证顺序
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
				if res != nil {
					objSlc = reflect.Append(objSlc, *res)
				}
			}
		}(part)
	}
	wg.Wait()

	endTime := time.Now()
	log.Infof("Jsons2StructsByRoutine rankId:%s,all:%d,notfound:%d,final:%d;total:%.4f\n",
		"", len(jsons), len(jsons)-objSlc.Len(), objSlc.Len(), 
		endTime.Sub(startTime).Seconds())
	return &objSlc, err
}

func (self *CachePikaModule) Jsons2Structs(jsons []interface{}, obj interface{}) (*reflect.Value, error) {
	var abtest = abtest.GetAbTest("", 0)
	if self.ctx != nil {
		abtest = self.ctx.GetAbTest()
	}
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
	log.Infof("UnmarshalKey:%s,rankId:%s,all:%d,notfound:%d,final:%d;total:%.4f,read:%.4f,json:%.4f\n",
		keyFormater, "", len(ids), len(ids)-objs.Len(), objs.Len(), 
		endTime.Sub(startTime).Seconds(),
		startJsonTime.Sub(startTime).Seconds(), endTime.Sub(startJsonTime).Seconds())
	return objs, err
}


// 单线程转化json为struct Map.去除空值，不保证顺序
func (self *CachePikaModule) Jsons2StructsMapBySingle(ids []int64, jsons []interface{}, obj interface{}) (*reflect.Value, error) {
	startTime := time.Now()
	objLen := len(jsons)
	var keyObj int64 = 0
	objKeyType := reflect.TypeOf(keyObj)
	objEleType := reflect.TypeOf(obj)
	objMapType := reflect.MapOf(objKeyType, objEleType)
	objMap := reflect.MakeMapWithSize(objMapType, objLen)
	ress, err := self.jsonsToValues(jsons, objEleType)
	for i, res := range ress {
		if res != nil {
			objMap.SetMapIndex(reflect.ValueOf(ids[i]), *res)
		}
	}
	finalLen := objMap.Len()
	endTime := time.Now()
	log.Infof("Jsons2StructsMapBySingle rankId:%s,all:%d,notfound:%d,final:%d;total:%.4f\n",
		"", len(jsons), len(jsons)-finalLen, finalLen, endTime.Sub(startTime).Seconds())
	return &objMap, err
}


// 多线程转化json为struct.去除空值，不保证顺序
func (self *CachePikaModule) Jsons2StructsMapByRoutine(ids []int64, jsons []interface{}, obj interface{}, partLen int) (*reflect.Value, error) {
	startTime := time.Now()
	objLen := len(jsons)

	// var keyObj int64 = 0
	objKeyType := reflect.TypeOf(ids).Elem()
	objEleType := reflect.TypeOf(obj)
	objMapType := reflect.MapOf(objKeyType, objEleType)
	objMap := reflect.MakeMapWithSize(objMapType, objLen)
	var err error

	var lockHandler = &sync.RWMutex{}
	wg := new(sync.WaitGroup)
	for _, part := range utils.SplitPartIndex(objLen, partLen) {
		wg.Add(1)
		go func(ids []int64, part []interface{}) {
			defer wg.Done()
			partRes, partErr := self.jsonsToValues(part, objEleType)
			lockHandler.Lock()
			defer lockHandler.Unlock()
			if err == nil {
				err = partErr
			}
			for j, res := range partRes {
				if res != nil {
					objMap.SetMapIndex(reflect.ValueOf(ids[j]), *res)
				}
			}
		}(ids[part.Start: part.End], jsons[part.Start: part.End])
	}
	wg.Wait()
	finalLen := objMap.Len()
	endTime := time.Now()
	log.Infof("Jsons2StructsMapByRoutine rankId:%s,all:%d,notfound:%d,final:%d;total:%.4f\n",
		"", len(jsons), len(jsons)-finalLen, finalLen, 
		endTime.Sub(startTime).Seconds())
	return &objMap, err
}

func (self *CachePikaModule) Jsons2StructsMap(ids []int64, jsons []interface{}, obj interface{}) (*reflect.Value, error) {
	var abtest = abtest.GetAbTest("", 0)
	if self.ctx != nil {
		abtest = self.ctx.GetAbTest()
	}
	threshold := abtest.GetInt("redis.json.map.thread.threshold", 200)
	if len(jsons) > threshold {
		jobs := abtest.GetInt("redis.json.map.thread.jobs", 4)
		return self.Jsons2StructsMapByRoutine(ids, jsons, obj, jobs)
	} else {
		return self.Jsons2StructsMapBySingle(ids, jsons, obj)
	}
}

func (self *CachePikaModule) MGetStructsMap(obj interface{}, ids []int64, keyFormater string, cacheTime int64, cacheNilTime int64) (*reflect.Value, error) {
	startTime := time.Now()
	ress, err := self.MGetSet(ids, keyFormater, cacheTime, cacheNilTime)
	startJsonTime := time.Now()
	objs, err := self.Jsons2StructsMap(ids, ress, obj)
	objs.MapKeys()
	endTime := time.Now()
	log.Infof("MGetStructsMap:%s,rankId:%s,all:%d,notfound:%d,final:%d;total:%.4f,read:%.4f,json:%.4f\n",
		keyFormater, "", len(ids), len(ids)-objs.Len(), objs.Len(), 
		endTime.Sub(startTime).Seconds(),
		startJsonTime.Sub(startTime).Seconds(), endTime.Sub(startJsonTime).Seconds())
	return objs, err
}

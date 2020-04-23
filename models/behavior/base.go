package behavior

import (
	"fmt"
	"math"
	"rela_recommend/cache"
	"rela_recommend/algo"
	"rela_recommend/utils"
	"rela_recommend/models/redis"
)

type BehaviorItemLog struct {
	DataId			int64			`json:"data_id"`
	UserId			int64			`json:"user_id"`
	LastTime		float64			`json:"last_time"`
}

type Behavior struct {
	Count			float64				`json:"count"`
	LastTime		float64				`json:"last_time"`
	LastList		[]BehaviorItemLog	`json:"last_list"`		// 最后操作列表
}

// 获取最后操作dataids列表
func(self *Behavior) GetLastDataIds() []int64 {
	ids := utils.SetInt64{}
	for _, log := range self.LastList {
		ids.Append(log.DataId)
	}
	return ids.ToList()
}

// 获取最后操作userids列表
func(self *Behavior) GetLastUserIds() []int64 {
	ids := utils.SetInt64{}
	for _, log := range self.LastList {
		ids.Append(log.UserId)
	}
	return ids.ToList()
}



// 合并行为
func MergeBehaviors(behaviors ...*Behavior) *Behavior {
	res := &Behavior{}
	for _, behavior := range behaviors {
		if behavior != nil {
			res.Count += behavior.Count
			res.LastTime = math.Max(res.LastTime, behavior.LastTime)
			res.LastList = append(res.LastList, behavior.LastList...)
		}
	}
	return res
}


type UserBehavior struct {
	CacheTime      			float64 				`json:"cache_time"`		// 缓存时间
	LastTime				float64 				`json:"last_time"`		// 最后动作时间
	Count					float64					`json:"count"`			// 触发动作次数
	BehaviorMap				map[string]*Behavior	`json:"behavior_map"`	// 各页面行为Map
}

func(self *UserBehavior) Get(name string) *Behavior {
	if self.BehaviorMap != nil {
		return self.BehaviorMap[name]
	}
	return nil
}

func(self *UserBehavior) Gets(names ...string) *Behavior {
	res := &Behavior{}
	if self.BehaviorMap != nil {
		for _, name := range names {
			if behavior, ok := self.BehaviorMap[name]; ok && behavior != nil {
				res.Count += behavior.Count
				res.LastTime = math.Max(res.LastTime, behavior.LastTime)
			}
		}
		return res
	}
	return res
}


type BehaviorCacheModule struct {
	redis.CachePikaModule
	ctx 	algo.IContext
}

// 读取user相关行为
func (self *BehaviorCacheModule) QueryUserBehaviorMap(module string, userId int64, ids []int64) (map[int64]*UserBehavior, error) {
	keyFormatter := fmt.Sprintf("behavior:%s:%d:%%d", module, userId)
	ress, err := self.MGetStructsMap(&UserBehavior{}, ids, keyFormatter, 0, 0)
	objs := ress.Interface().(map[int64]*UserBehavior)
	return objs, err
}

// 读取item相关行为
func (self *BehaviorCacheModule) QueryItemBehaviorMap(module string, ids []int64) (map[int64]*UserBehavior, error) {
	keyFormatter := fmt.Sprintf("behavior:%s:%%d", module)
	ress, err := self.MGetStructsMap(&UserBehavior{}, ids, keyFormatter, 0, 0)
	objs := ress.Interface().(map[int64]*UserBehavior)
	return objs, err
}

func NewBehaviorCacheModule(ctx algo.IContext, cache *cache.Cache) *BehaviorCacheModule {
	cachePika := redis.NewCachePikaModule(ctx, *cache)
	return &BehaviorCacheModule{CachePikaModule: *cachePika, ctx: ctx}
}

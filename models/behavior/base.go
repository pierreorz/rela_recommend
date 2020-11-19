package behavior

import (
	"errors"
	"fmt"
	"math"
	"rela_recommend/cache"
	"rela_recommend/models/redis"
	"rela_recommend/service/abtest"
	"rela_recommend/utils"
	"sort"
)

type BehaviorItemLog struct {
	DataId   int64   `json:"data_id"`
	UserId   int64   `json:"user_id"`
	LastTime float64 `json:"last_time"`
}

///                    *********************** tags
type BehaviorTag struct {
	Id       int64   `json:"id"`
	Category string  `json:"category"`
	Name     string  `json:"name"`
	Count    float64 `json:"count"`
	LastTime float64 `json:"last_time"`
}

func (self *BehaviorTag) Merge(other *BehaviorTag) *BehaviorTag {
	self.Id = other.Id
	self.Category = other.Category
	self.Name = other.Name
	self.Count += other.Count
	self.LastTime = math.Max(self.LastTime, other.LastTime)
	return self
}

// 排序
type behaviorTagSorter []*BehaviorTag

func (a behaviorTagSorter) Len() int      { return len(a) }
func (a behaviorTagSorter) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a behaviorTagSorter) Less(i, j int) bool { // 按照 Count , lastTime, Name 倒序
	if a[i].Count == a[j].Count {
		if a[i].LastTime == a[j].LastTime {
			return a[i].Id > a[j].Id
		} else {
			return a[i].LastTime > a[j].LastTime
		}
	} else {
		return a[i].Count > a[j].Count
	}
}

type Behavior struct {
	Count    float64                 `json:"count"`
	LastTime float64                 `json:"last_time"`
	LastList []BehaviorItemLog       `json:"last_list"` // 最后操作列表
	CountMap map[string]*BehaviorTag `json:"count_map"` // 类别对应的标签
}

// 获取最后操作dataids列表
func (self *Behavior) GetLastDataIds() []int64 {
	ids := utils.SetInt64{}
	for _, log := range self.LastList {
		ids.Append(log.DataId)
	}
	return ids.ToList()
}

// 获取最后操作userids列表
func (self *Behavior) GetLastUserIds() []int64 {
	ids := utils.SetInt64{}
	for _, log := range self.LastList {
		ids.Append(log.UserId)
	}
	return ids.ToList()
}

// 获取top n的map，返回去除前缀
func (self *Behavior) GetTopCountTags(category string, n int) []*BehaviorTag {
	var res = behaviorTagSorter{}
	for key, tag := range self.CountMap {
		if tag.Category == category {
			res = append(res, self.CountMap[key])
		}
	}
	sort.Sort(res)
	if n < len(res) {
		res = res[:n]
	}
	return res
}

func (self *Behavior) GetTopCountTagsMap(category string, n int) map[int64]*BehaviorTag {
	tagMap := map[int64]*BehaviorTag{}
	tags := self.GetTopCountTags(category, n)
	for i, tag := range tags {
		tagMap[tag.Id] = tags[i]
	}
	return tagMap
}

func (self *Behavior) Merge(other *Behavior) *Behavior {
	if self.CountMap == nil {
		self.CountMap = map[string]*BehaviorTag{}
	}
	if other != nil {
		self.Count += other.Count
		self.LastTime = math.Max(self.LastTime, other.LastTime)
		self.LastList = append(self.LastList, other.LastList...)
		for categoryName, tag := range other.CountMap {
			if current, ok := self.CountMap[categoryName]; ok {
				self.CountMap[categoryName] = current.Merge(tag)
			} else {
				self.CountMap[categoryName] = other.CountMap[categoryName]
			}
		}
	}
	return self
}

// 合并行为
func MergeBehaviors(behaviors ...*Behavior) *Behavior {
	res := &Behavior{}
	for _, behavior := range behaviors {
		res = res.Merge(behavior)
	}
	return res
}

type UserBehavior struct {
	CacheTime   float64              `json:"cache_time"`   // 缓存时间
	LastTime    float64              `json:"last_time"`    // 最后动作时间
	Count       float64              `json:"count"`        // 触发动作次数
	BehaviorMap map[string]*Behavior `json:"behavior_map"` // 各页面行为Map
}

func (self *UserBehavior) Get(name string) *Behavior {
	if self.BehaviorMap != nil {
		return self.BehaviorMap[name]
	}
	return nil
}

func (self *UserBehavior) Gets(names ...string) *Behavior {
	res := &Behavior{}
	if self.BehaviorMap != nil {
		var behaviors = []*Behavior{}
		for _, name := range names {
			if behavior, ok := self.BehaviorMap[name]; ok && behavior != nil {
				behaviors = append(behaviors, behavior)
			}
		}
		res = MergeBehaviors(behaviors...)
	}
	return res
}

type BehaviorCacheModule struct {
	redis.CachePikaModule
	ctx abtest.IAbTestAble
}

// 读取user item相关行为
func (self *BehaviorCacheModule) QueryUserItemBehaviorMap(module string, userId int64, ids []int64) (map[int64]*UserBehavior, error) {
	keyFormatter := fmt.Sprintf("behavior:%s:%d:%%d.gz", module, userId)
	ress, err := self.MGetStructsMap(&UserBehavior{}, ids, keyFormatter, 0, 0)
	objs := ress.Interface().(map[int64]*UserBehavior)
	return objs, err
}

// 读取item相关行为
func (self *BehaviorCacheModule) QueryItemBehaviorMap(module string, ids []int64) (map[int64]*UserBehavior, error) {
	keyFormatter := fmt.Sprintf("behavior:%s:item:%%d.gz", module)
	ress, err := self.MGetStructsMap(&UserBehavior{}, ids, keyFormatter, 0, 0)
	objs := ress.Interface().(map[int64]*UserBehavior)
	return objs, err
}

// 读取user相关行为
func (self *BehaviorCacheModule) QueryUserBehaviorMap(module string, ids []int64) (map[int64]*UserBehavior, error) {
	keyFormatter := fmt.Sprintf("behavior:%s:user:%%d.gz", module)
	ress, err := self.MGetStructsMap(&UserBehavior{}, ids, keyFormatter, 0, 0)
	objs := ress.Interface().(map[int64]*UserBehavior)
	return objs, err
}

func NewBehaviorCacheModule(ctx abtest.IAbTestAble, cache *cache.Cache) *BehaviorCacheModule {
	cachePika := redis.NewCachePikaModule(ctx, *cache)
	return &BehaviorCacheModule{CachePikaModule: *cachePika, ctx: ctx}
}

// *************************************** 内容行为分数
type DataBehaviorScore struct {
	DataId int64   `json:"dataId"` // 数据id
	Score  float64 `json:"score"`  // 得分
}

type DataBehaviorTopList struct {
	Data []DataBehaviorScore `json:"data"` // 热门列表
}

func (self *DataBehaviorTopList) GetTopIds(n int) []int64 {
	res := []int64{}
	for i, topItem := range self.Data {
		if i >= n {
			break
		}
		res = append(res, topItem.DataId)
	}
	return res
}

func (self *BehaviorCacheModule) QueryDataBehaviorTop(module string) (*DataBehaviorTopList, error) {
	if self.ctx != nil {
		topDataKey := self.ctx.GetAbTest().GetString("behavior_data_top_key", "behavior:item:%s:top")
		keyFormatter := fmt.Sprintf(topDataKey, module)
		topList := &DataBehaviorTopList{}
		err := self.GetStruct(keyFormatter, topList)
		return topList, err
	} else {
		return nil, errors.New("context is nil")
	}
}

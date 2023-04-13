package behavior

import (
	"errors"
	"fmt"
	"math"
	"rela_recommend/factory"
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

///    *********picture tags
type PictureTag struct {
	LastTime float64 `json:"last_time"`
	Name     string  `json:"name"`
	Count    float64 `json:"count"`
}

type MomTag struct {
	LastTime float64 `json:"last_time"`
	Id       string  `json:"id"`
	Count    float64 `json:"count"`
}

func (self *BehaviorTag) Merge(other *BehaviorTag) *BehaviorTag {
	self.Id = other.Id
	self.Category = other.Category
	self.Name = other.Name
	self.Count += other.Count
	self.LastTime = math.Max(self.LastTime, other.LastTime)
	return self
}

func (self *PictureTag) Merge(other *PictureTag) *PictureTag {
	self.LastTime = math.Max(other.LastTime, self.LastTime)
	self.Name = other.Name
	self.Count += other.Count
	return self
}

func (self *MomTag) Merge(other *MomTag) *MomTag {
	self.LastTime = math.Max(other.LastTime, self.LastTime)
	self.Id = other.Id
	self.Count += other.Count
	return self
}

// 排序
type pictureTagSorter []*PictureTag
type behaviorTagSorter []*BehaviorTag
type momTagSorter []*MomTag

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

func (a pictureTagSorter) Len() int      { return len(a) }
func (a pictureTagSorter) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a pictureTagSorter) Less(i, j int) bool { // 按照 Count , lastTime, Name 倒序
	if a[i].Count == a[j].Count {
		if a[i].LastTime == a[j].LastTime {
			return a[i].Name > a[j].Name
		} else {
			return a[i].LastTime > a[j].LastTime
		}
	} else {
		return a[i].Count > a[j].Count
	}
}

func (a momTagSorter) Len() int      { return len(a) }
func (a momTagSorter) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a momTagSorter) Less(i, j int) bool { // 按照 Count , lastTime, Name 倒序
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
	Count      float64                 `json:"count"`
	LastTime   float64                 `json:"last_time"`
	LastList   []BehaviorItemLog       `json:"last_list"` // 最后操作列表
	CountMap   map[string]*BehaviorTag `json:"count_map"` // 类别对应的标签
	PictureMap map[string]*PictureTag  `json:"picture_map"`
	MomTagMap  map[string]*MomTag      `json:"momTag_map"`
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

//获取最后一次广告曝光数据
func (self *Behavior) GetLastAdIds() []int64 {
	ids := utils.SetInt64{}
	for _, log := range self.LastList {
		ids.Append(log.DataId)
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

func (self *Behavior) GetTopCountPictureTags(n int) []*PictureTag {
	var res = pictureTagSorter{}
	for key, tag := range self.PictureMap {
		if LabelConvert(tag.Name) != "" {
			res = append(res, self.PictureMap[key])
		}
	}
	sort.Sort(res)
	if n < len(res) {
		res = res[:n]
	}
	return res
}

func (self *Behavior) GetTopCountMomTags(n int) []*MomTag {
	var res = momTagSorter{}
	for key, tag := range self.MomTagMap {
		if len(tag.Id) > 0 {
			res = append(res, self.MomTagMap[key])
		}
	}
	sort.Sort(res)
	if n < len(res) {
		res = res[:n]
	}
	return res
}

func LabelConvert(label string) string {
	if utils.StringContains(label, []string{"burudongwu"}) {
		return "pet"
	}
	if utils.StringContains(label, []string{"dongman"}) {
		return "dongman"
	}
	if utils.StringContains(label, []string{"biaoqingbao"}) {
		return "biaoqingbao"
	}
	if utils.StringContains(label, []string{"jianshen"}) {
		return "yundong"
	}
	if utils.StringContains(label, []string{"youxi"}) {
		return "youxi"
	}
	if utils.StringContains(label, []string{"shaoshumingzufushi", "xiaofu", "JKzhifu", "qizhi"}) {
		return "shishang"
	}
	if utils.StringContains(label, []string{"meishi"}) {
		return "meishi"
	}
	if utils.StringContains(label, []string{"jiejing", "ziranfengguang"}) {
		return "fengjing"
	}
	return ""
}
func (self *Behavior) GetTopCountPictureTagsMap(n int) map[string]*PictureTag {
	tagMap := map[string]*PictureTag{}
	tags := self.GetTopCountPictureTags(n)
	for i, tag := range tags {
		tagMap[tag.Name] = tags[i]
	}
	return tagMap
}

func (self *Behavior) GetTopCountTagsMap(category string, n int) map[int64]*BehaviorTag {
	tagMap := map[int64]*BehaviorTag{}
	tags := self.GetTopCountTags(category, n)
	for i, tag := range tags {
		tagMap[tag.Id] = tags[i]
	}
	return tagMap
}

func (self *Behavior) GetTopCountMomTagsMap(n int) map[string]*MomTag {
	tagMap := map[string]*MomTag{}
	tags := self.GetTopCountMomTags(n)
	for i, tag := range tags {
		tagMap[tag.Id] = tags[i]
	}
	return tagMap
}

func (self *Behavior) Merge(other *Behavior) *Behavior {
	if self == nil {
		return other
	}
	if self.CountMap == nil {
		self.CountMap = map[string]*BehaviorTag{}
	}
	if self.PictureMap == nil {
		self.PictureMap = map[string]*PictureTag{}
	}
	if self.MomTagMap == nil {
		self.MomTagMap = map[string]*MomTag{}
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
		for label, pictureTag := range other.PictureMap {
			if current, ok := self.PictureMap[label]; ok {
				self.PictureMap[label] = current.Merge(pictureTag)
			} else {
				self.PictureMap[label] = other.PictureMap[label]
			}
		}
		for label, momTag := range other.MomTagMap {
			if current, ok := self.MomTagMap[label]; ok {
				self.MomTagMap[label] = current.Merge(momTag)
			} else {
				self.MomTagMap[label] = other.MomTagMap[label]
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

func (self *UserBehavior) Merge(other *UserBehavior) *UserBehavior {
	if self == nil {
		return other
	}
	if self.BehaviorMap == nil {
		self.BehaviorMap = map[string]*Behavior{}
	}
	if other != nil {
		self.Count += other.Count
		self.LastTime = math.Max(self.LastTime, other.LastTime)
		self.CacheTime = math.Max(self.CacheTime, other.CacheTime)
		for pageName, behaviors := range other.BehaviorMap {
			if current, ok := self.BehaviorMap[pageName]; ok {
				self.BehaviorMap[pageName] = current.Merge(behaviors)
			} else {
				self.BehaviorMap[pageName] = other.BehaviorMap[pageName]
			}
		}
	}
	return self
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

// QueryBeenUserItemBehaviorMap 读取user item相关别互动行为，
func (self *BehaviorCacheModule) QueryBeenUserItemBehaviorMap(module string, userId int64, ids []int64) (map[int64]*UserBehavior, error) {
	keyFormatter := fmt.Sprintf("behavior:%s:%%d:%d.gz", module, userId)
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

func NewBehaviorCacheModule(ctx abtest.IAbTestAble) *BehaviorCacheModule {
	cache := factory.CacheBehaviorRdsBackup
	cachePika := redis.NewCachePikaModule(ctx, cache)
	return &BehaviorCacheModule{CachePikaModule: *cachePika, ctx: ctx}
}

// *************************************** 内容行为分数
type DataBehaviorScore struct {
	DataId    int64   `json:"dataId"` // 数据id
	Exposure  int     `json:"exposure"`
	Like      int     `json:"like"`
	Comment   int     `json:"comment"`
	LastTime  float64 `json:"lastTime"`
	FirstTime float64 `json:"firstTime"`
	Share     int     `json:"share"`
	Follow    int     `json:"follow"`
	Score     float64 `json:"score"` // 得分
}

type DataBehaviorTopList struct {
	Data []DataBehaviorScore `json:"data"` // 热门列表
}

func (self *DataBehaviorTopList) GetTopIdsV2(score float64) []int64 {
	res := []int64{}
	for _, topItem := range self.Data {
		if topItem.Score > score && topItem.Comment <= 2 {
			res = append(res, topItem.DataId)
		}
	}
	return res
}

func (self *DataBehaviorTopList) GetTopIds(n int) []int64 {
	res := []int64{}
	count := 0
	for _, topItem := range self.Data {
		if count >= n {
			break
		}
		if topItem.Comment < 2 {
			res = append(res, topItem.DataId)
			count += 1
		}
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

func (self *BehaviorCacheModule) QueryMateMomUserData(userId []int64) map[int64][]int64 {
	userProfileUserIds := userId
	userMomMap := make(map[int64][]int64)
	realtimes, realtimeErr := self.QueryUserBehaviorMap("moment", userProfileUserIds)
	var userList []int64
	if realtimeErr == nil {
		for userId, momProfile := range realtimes {
			if momProfile != nil {
				userList = append(userList, userId)
				countMap := momProfile.BehaviorMap["moment.recommend:exposure"]
				if countMap != nil {
					tagMap := countMap.CountMap
					if tagMap != nil {
						for _, v := range tagMap {
							userMomMap[v.Id] = userList
						}
					}
				}
			}
		}
		return userMomMap
	}
	return userMomMap
}

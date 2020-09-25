package redis

import (
	"fmt"
	"rela_recommend/algo"
	"rela_recommend/cache"
	"rela_recommend/utils"
)

// 直播业务的画像
type LiveProfile struct {
	UserId int64 `json:"user_id"`
	// 观看直播间时间的 embedding
	LiveViewUserEmbedding []float32 `json:"live_view5_user"` // 用户端
	LiveViewLiveEmbedding []float32 `json:"live_view5_live"` // 主播端
}

type LiveCacheModule struct {
	CachePikaModule
}

func NewLiveCacheModule(ctx algo.IContext, cache *cache.Cache, store *cache.Cache) *LiveCacheModule {
	return &LiveCacheModule{CachePikaModule{ctx: ctx, cache: *cache, store: *store}}
}

// 读取直播相关用户画像
func (self *LiveCacheModule) QueryLiveProfileByUserIds(ids []int64) ([]LiveProfile, error) {
	keyFormatter := "live_profile:%d"
	ress, err := self.MGetStructs(LiveProfile{}, ids, keyFormatter, 24*60*60, 60*60*1)
	objs := ress.Interface().([]LiveProfile)
	return objs, err
}

func (self *LiveCacheModule) QueryLiveProfileByUserAndUsersMap(userId int64, userIds []int64) (*LiveProfile, map[int64]*LiveProfile, error) {
	allIds := append(userIds, userId)
	users, err := self.QueryLiveProfileByUserIds(allIds)
	var resUser *LiveProfile
	var resUsers = map[int64]*LiveProfile{}
	if err == nil {
		for i, user := range users {
			if user.UserId == userId {
				resUser = &users[i]
			} else {
				resUsers[user.UserId] = &users[i]
			}
		}
	}
	return resUser, resUsers, err
}

// 从缓存中获取以逗号分割的字符串，并转化成int64. 如 keys11  1,2,3,4,5
func (self *LiveCacheModule) GetInt64List(id int64, keyFormatter string) ([]int64, error) {
	res, err := self.GetSet(fmt.Sprintf(keyFormatter, id), 24*60*60, 1*60*60)
	if err == nil {
		return utils.GetInt64s(utils.GetString(res)), nil
	}
	return nil, err
}

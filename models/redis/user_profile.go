package redis

import(
	"time"
	"errors"
	"encoding/json"
	"rela_recommend/log"
	"rela_recommend/cache"
)


type UserProfile struct {
	UserId 	int64		`json:"id"`
	// 观看直播间时间的 embedding
	LiveViewUserEmbedding []float32		`json:"live_view_user"`		// 用户端
	LiveViewLiveEmbedding []float32		`json:"live_view_live"`		// 主播端
}


type UserProfileModule struct {
	cache cache.Cache
	store cache.Cache
}

func NewUserProfileModule(cache *cache.Cache, store *cache.Cache) *UserProfileModule {
	return &UserProfileModule{cache: *cache, store: *store}
}

func (self *UserProfileModule) QueryByUserIds(userIds []int64) ([]UserProfile, error) {
	startTime := time.Now()
	cacheModule := &CachePikaModule{cache: self.cache, store: self.store}
	keyFormatter := "algo_user_profile:%d"
	ress, err := cacheModule.MGetSet(userIds, keyFormatter, 24 * 60 * 60)
	startJsonTime := time.Now()
	users := make([]UserProfile, 0)
	for i, res := range ress {
		if res != nil {
			var user UserProfile
			bs, ok := res.([]byte)
			if ok {
				if err := json.Unmarshal(bs, &user); err == nil {
					users = append(users, user)
				} else {
					log.Warn(keyFormatter, userIds[i], err.Error())
				}
			} else {
				log.Warn(keyFormatter, userIds[i], err.Error())
			}
		}
	}
	endTime := time.Now()
	log.Infof("UnmarshalKey:%s,all:%d,cache:%d,final:%d;total:%.3f,read:%.3f,json:%.3f\n",
		keyFormatter, len(userIds), len(userIds)-len(users), len(users), 
		endTime.Sub(startTime).Seconds(),
		startJsonTime.Sub(startTime).Seconds(), endTime.Sub(startJsonTime).Seconds())
	return users, err
}


func (self *UserProfileModule) QueryByUserAndUsers(userId int64, userIds []int64) (UserProfile, []UserProfile, error) {
	allIds := append(userIds, userId)
	users, err := self.QueryByUserIds(allIds)
	var resUser UserProfile
	var resUsers []UserProfile
	if err == nil {
		for i, user := range users {
			if user.UserId == userId {
				resUser = user
				resUsers = append(users[:i], users[i+1:]...)
				// users i后面的内容向前移动了一位，内容发上了改变，谨慎使用
				break
			}
		}
		if resUser.UserId == 0 {
			err = errors.New("user is nil")
		}
	}
	return resUser, resUsers, err
}

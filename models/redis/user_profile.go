package redis

import(
	"time"
	"errors"
	"encoding/json"
	"rela_recommend/log"
	"rela_recommend/cache"
	"rela_recommend/utils"
)


// 直播业务的画像
type LiveProfile struct {
	UserId 	int64		`json:"user_id"`
	// 观看直播间时间的 embedding
	LiveViewUserEmbedding []float32		`json:"live_view5_user"`		// 用户端
	LiveViewLiveEmbedding []float32		`json:"live_view5_live"`		// 主播端
}


type UserProfileModule struct {
	cache cache.Cache
	store cache.Cache
}

func NewUserProfileModule(cache *cache.Cache, store *cache.Cache) *UserProfileModule {
	return &UserProfileModule{cache: *cache, store: *store}
}

func (self *UserProfileModule) QueryLiveProfileByUserIds(userIds []int64) ([]LiveProfile, error) {
	startTime := time.Now()
	cacheModule := &CachePikaModule{cache: self.cache, store: self.store}
	keyFormatter := "live_profile:%d"
	ress, err := cacheModule.MGetSet(userIds, keyFormatter, 24 * 60 * 60, 60 * 30)
	startJsonTime := time.Now()
	users := make([]LiveProfile, 0)
	for i, res := range ress {
		if res != nil {
			var user LiveProfile
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


func (self *UserProfileModule) QueryLiveProfileByUserAndUsers(userId int64, userIds []int64) (LiveProfile, []LiveProfile, error) {
	allIds := append(userIds, userId)
	users, err := self.QueryLiveProfileByUserIds(allIds)
	var resUser LiveProfile
	var resUsers []LiveProfile
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
			err = errors.New("LiveProfile user is nil" + utils.GetString(userId))
		}
	}
	return resUser, resUsers, err
}

package redis

import(
	"time"
	"errors"
	"encoding/json"
	"rela_recommend/log"
	"rela_recommend/cache"
	"rela_recommend/utils"
	"fmt"
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

// 读取直播相关用户画像
func (self *UserProfileModule) QueryLiveProfileByUserIds(userIds []int64) ([]LiveProfile, error) {
	startTime := time.Now()
	cacheModule := &CachePikaModule{cache: self.cache, store: self.store}
	keyFormatter := "live_profile:%d"
	ress, err := cacheModule.MGetSet(userIds, keyFormatter, 24 * 60 * 60, 60 * 60 * 1)
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
	log.Infof("UnmarshalKey:%s,all:%d,notfound:%d,final:%d;total:%.4f,read:%.4f,json:%.4f\n",
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
			err = errors.New("LiveProfile user is nil " + utils.GetString(userId))
		}
	}
	return resUser, resUsers, err
}

// 从缓存中获取以逗号分割的字符串，并转化成int64. 如 keys11  1,2,3,4,5
func (self *UserProfileModule) GetInt64List(id int64, keyFormatter string) ([]int64, error) {
	cacheModule := &CachePikaModule{cache: self.cache, store: self.store}
	res, err := cacheModule.GetSet(id, keyFormatter, 24 * 60 * 60, 1 * 60 * 60)
	fmt.Println(res)
	if err == nil {
		return utils.GetInt64s(utils.GetString(res)), nil
	}
	return nil, err
}

package pika

import (
	"encoding/json"
	"errors"
	"rela_recommend/factory"
	"rela_recommend/log"
	"rela_recommend/utils"
	"rela_recommend/cache"
	"time"
	// "strings"
)

type JsonTime struct {
	time.Time
}
func (p *JsonTime) UnmarshalJSON(data []byte) error {
	local, err := time.ParseInLocation("\"2006-01-02T15:04:05.000+0000\"", string(data), time.Local)
	*p = JsonTime{Time: local}
	return err
}
func (c JsonTime) MarshalJSON() ([]byte, error) {
	data := make([]byte, 0)
	data = append(data, '"')
	data = time.Time(c.Time).AppendFormat(data, "2006-01-02 15:04:05.000+0000")
	data = append(data, '"')
	return data, nil
}

type UserProfileModule struct {
	cacheCluster *cache.Cache
	storeCluster *cache.Cache
}

func NewUserProfileModule(cache *cache.Cache, store *cache.Cache) *UserProfileModule {
	return &UserProfileModule{cacheCluster: cache, storeCluster: store}
}

type Location struct {
	Lat 	float64		`bson:"lat"`
	Lon 	float64		`bson:"lon"`
}

type UserProfile struct {
	UserId         int64    `bson:"id"`         // 用户ID
	Location       Location `bson:"location"`            //地理位置
	Avatar         string   `bson:"avatar"`         // 头像
	IsVip          int      `bson:"isVip"`          // 是否是vip
	LastUpdateTime int64    `bson:"lastUpdateTime"` //最后在线时间
	MomentsCount   int      `bson:"momentsCount"`   // 日志数
	NewImageCount  int      `bson:"newImageCount"`
	RoleName       string   `bson:"roleName"`
	UserImageCount int      `bson:"userImageCount"`
	WantRole       string   `bson:"wantRole"`

	Affection  int       `bson:"affection"`
	Age        int       `bson:"age"`
	Height     int       `bson:"height"`
	Weight     int       `bson:"weight"`
	Ratio      int       `bson:"ratio"`
	CreateTime time.Time `bson:"createTime"`
	Horoscope  string    `bson:"horoscope"`

	JsonRoleLike map[string]float32	`bson:"jsonRoleLike"`
	JsonAffeLike map[string]float32	`bson:"jsonAffeLike"`
}

func (this *UserProfileModule) QueryByUserIds(userIds []int64) ([]UserProfile, error) {
	var cacheKeyPre = "active_location_location:"
	var storeKeyPre = "active_location_location:"
	auls := make([]UserProfile, 0)
	var startTime = time.Now()
	var cacheKeysMap = map[int64]string{}
	var cacheKeys = make([]string, 0)
	for _, id := range userIds {    // 构造缓存keys
		cacheKey := cacheKeyPre + utils.GetString(id)
		cacheKeysMap[id] = cacheKey
		cacheKeys = append(cacheKeys, cacheKey)
	}
	var startRedisTime = time.Now()  // 开始读取缓存
	userStrs, err := factory.CacheCluster.Mget(cacheKeys)
	if err != nil {
		log.Error(err.Error())
	}

	var startRedisResTime = time.Now()  // 开始解析缓存结果
	usersMap := map[int64]UserProfile{}
	var notFoundUserIds = make([]int64, 0)
	for i, userId := range userIds {
		userRes := userStrs[i]
		if userRes == nil {
			notFoundUserIds = append(notFoundUserIds, userId)
			continue
		}
		var user UserProfile
		userStr := utils.GetString(userRes)
		// userStr = strings.Replace(userStr, "+0000\"", "Z\"", -1)
		if err := json.Unmarshal(([]byte)(userStr), &user); err != nil {
			notFoundUserIds = append(notFoundUserIds, userId)
			log.Error(err.Error())
		} else {
			usersMap[userId] = user
		}
	}
	var startNFTime = time.Now()
	var startMongoTime = time.Now()
	var start2RedisResTime = time.Now()
	var start2RedisTime = time.Now()
	if len(notFoundUserIds) > 0 {
		var storeKeys = make([]string, 0)
		for _, id := range notFoundUserIds {
			storeKeys = append(storeKeys, storeKeyPre + utils.GetString(id))
		}
		startMongoTime = time.Now()  // 开始读取持久化存储
		storeUserStrs, err := factory.PikaCluster.Mget(storeKeys)
		if err != nil {
			log.Error(err.Error())
		}
		start2RedisResTime = time.Now()  // 开始解析持久化存储结果
		toCacheMap := map[string]interface{}{}
		for i, userId := range notFoundUserIds {
			// storeKey := storeKeys[i]
			userRes := storeUserStrs[i]
			if userRes == nil {
				continue
			}
			var user UserProfile
			userStr := utils.GetString(userRes)
			// userStr = strings.Replace(userStr, "+0000\"", "Z\"", -1)
			if err := json.Unmarshal(([]byte)(userStr), &user); err != nil {
				log.Error(err.Error())
			} else {
				usersMap[userId] = user

				cacheUserKey, _ := cacheKeysMap[userId]
				toCacheMap[cacheUserKey] = userStr
			}
		}

		start2RedisTime = time.Now()  // 开始将持久化结果写入缓存
		err = factory.CacheCluster.MsetEx(toCacheMap, 24 * 60 * 60)
		if err != nil {
			log.Error(err.Error())
		}
	}
	for _, userId := range userIds {
		user, found := usersMap[userId]
		if found && user.UserId == userId {
			auls = append(auls, user)
		}
	}
	var startLogTime = time.Now()
	log.Infof("QueryByUserIds,redis:%d,mongo:%d;total:%.3f,redisInit:%.3f,redis:%.3f,redisLoad:%.3f,notfound:%.3f,mongo:%.3f,2redisInit:%.3f,2redis:%.3f\n",
		len(userIds), len(auls),
		startLogTime.Sub(startTime).Seconds(), startRedisTime.Sub(startTime).Seconds(),
		startRedisResTime.Sub(startRedisTime).Seconds(), startNFTime.Sub(startRedisResTime).Seconds(),
		startMongoTime.Sub(startNFTime).Seconds(), start2RedisResTime.Sub(startMongoTime).Seconds(),
		start2RedisTime.Sub(start2RedisResTime).Seconds(), startLogTime.Sub(start2RedisTime).Seconds())
	return auls, err
}

func (this *UserProfileModule) QueryByUserAndUsers(userId int64, userIds []int64) (UserProfile, []UserProfile, error) {
	allIds := append(userIds, userId)
	users, err := this.QueryByUserIds(allIds)
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

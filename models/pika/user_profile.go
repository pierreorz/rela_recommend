package pika

import (
	"encoding/json"
	"errors"
	"rela_recommend/factory"
	"rela_recommend/log"
	"rela_recommend/utils"
	"rela_recommend/cache"
	"time"
	"strings"
)

type JsonTime struct {
	time.Time
}
func (p *JsonTime) UnmarshalJSON(data []byte) error {
	dataStr := string(data)
	if data != nil && dataStr!="null" && len(data) > 0 {
		var local time.Time
		var err error
		if strings.HasSuffix(dataStr, "+0000\"") {
			local, err = time.ParseInLocation("\"2006-01-02T15:04:05.000+0000\"", dataStr, time.Local)
		} else {
			(&local).UnmarshalJSON(data)
		}
		*p = JsonTime{Time: local}
		return err
	} else {
		return nil
	}
}
func (c *JsonTime) MarshalJSON() ([]byte, error) {
	if &c.Time != nil {
		data := make([]byte, 0)
		data = append(data, '"')
		data = time.Time(c.Time).AppendFormat(data, "2006-01-02 15:04:05.000+0000")
		data = append(data, '"')
		return data, nil
	} else {
		return nil, nil
	}
}

type UserProfileModule struct {
	cacheCluster *cache.Cache
	storeCluster *cache.Cache
}

func NewUserProfileModule(cache *cache.Cache, store *cache.Cache) *UserProfileModule {
	return &UserProfileModule{cacheCluster: cache, storeCluster: store}
}

type Location struct {
	Lat 	float64		`json:"lat"`
	Lon 	float64		`json:"lon"`
}

type UserProfile struct {
	UserId         int64    `json:"id"`         // 用户ID
	Location       Location `json:"location"`            //地理位置
	Avatar         string   `json:"avatar"`         // 头像
	IsVip          int      `json:"isVip"`          // 是否是vip
	LastUpdateTime int64    `json:"lastUpdateTime"` //最后在线时间
	MomentsCount   int      `json:"momentsCount"`   // 日志数
	NewImageCount  int      `json:"newImageCount"`
	RoleName       string   `json:"roleName"`
	UserImageCount int      `json:"userImageCount"`
	WantRole       string   `json:"wantRole"`

	Affection  int       `json:"affection"`
	Age        int       `json:"age"`
	Height     int       `json:"height"`
	Weight     int       `json:"weight"`
	Ratio      int       `json:"ratio"`
	CreateTime JsonTime `json:"createTime"`
	Horoscope  string    `json:"horoscope"`

	JsonRoleLike map[string]float32	`json:"jsonRoleLike"`
	JsonAffeLike map[string]float32	`json:"jsonAffeLike"`
}

func (this *UserProfileModule) QueryByUserIds(userIds []int64) ([]UserProfile, error) {
	var cacheKeyPre = "app_user_location:"
	var storeKeyPre = "app_user_location:"
	auls := make([]UserProfile, 0, len(userIds))
	usersMap := map[int64]UserProfile{}
	var startTime = time.Now()
	var cacheKeysMap = map[int64]string{}
	var cacheKeys = make([]string, 0)
	for _, id := range userIds {    // 构造缓存keys
		cacheKey := cacheKeyPre + utils.GetString(id)
		cacheKeysMap[id] = cacheKey
		cacheKeys = append(cacheKeys, cacheKey)
	}
	var startRedisTime = time.Now()  // 开始读取缓存
	var notFoundUserIds = make([]int64, 0)
	userStrs, err := factory.CacheCluster.Mget(cacheKeys)
	var startRedisResTime = time.Now()  // 开始解析缓存结果
	if err != nil {  // 读取缓存失败
		log.Error(err.Error())
		notFoundUserIds = userIds
	} else {
		for i, userId := range userIds {  
			userRes := userStrs[i]
			if userRes == nil {
				notFoundUserIds = append(notFoundUserIds, userId)
				continue
			}
			var user UserProfile
			userStr := utils.GetString(userRes)
			// log.Info(userStr)
			// userStr = strings.Replace(userStr, "+0000\"", "Z\"", -1)
			if err := json.Unmarshal(([]byte)(userStr), &user); err != nil {
				notFoundUserIds = append(notFoundUserIds, userId)
				log.Error(userId, err.Error())
			} else {
				usersMap[userId] = user
			}
		}
	}

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
		if err != nil {  // 读取持久化存储失败
			log.Error(err.Error())
			return nil, err
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
				log.Error(userId, err.Error())
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
		} else {
			log.Error(errors.New("can't found user " + utils.GetString(userId)))
		}
	}
	var startLogTime = time.Now()
	log.Infof("QueryByUserIds,all:%d,redis:%d,pika:%d,final:%d;total:%.3f,redisInit:%.3f,redis:%.3f,redisLoad:%.3f,pika:%.3f,pikaLoad:%.3f,2redis:%.3f\n",
		len(userIds), len(userIds)-len(notFoundUserIds), len(notFoundUserIds),len(auls),
		startLogTime.Sub(startTime).Seconds(), startRedisTime.Sub(startTime).Seconds(),
		startRedisResTime.Sub(startRedisTime).Seconds(), startMongoTime.Sub(startRedisResTime).Seconds(),
		start2RedisResTime.Sub(startMongoTime).Seconds(),
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

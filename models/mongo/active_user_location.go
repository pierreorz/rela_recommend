package mongo

import (
	"encoding/json"
	"errors"
	redis2 "github.com/chasex/redis-go-cluster"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"rela_recommend/factory"
	"rela_recommend/log"
	"rela_recommend/utils"
	"time"
)

type ActiveUserLocationModule struct {
	session *mgo.Session
}

func NewActiveUserLocationModule(session *mgo.Session) *ActiveUserLocationModule {
	return &ActiveUserLocationModule{session: session}
}

type Loc struct {
	Ty          string    `bson:"type"`        // default "Point"
	Coordinates []float64 `bson:"coordinates"` //经纬度
}

type ActiveUserLocation struct {
	UserId         int64  `bson:"userId"`         // 用户ID
	Loc            Loc    `bson:"loc"`            //地理位置
	Avatar         string `bson:"avatar"`         // 头像
	IsVip          int    `bson:"isVip"`          // 是否是vip
	LastUpdateTime int64  `bson:"lastUpdateTime"` //最后在线时间
	MomentsCount   int    `bson:"momentsCount"`   // 日志数
	NewImageCount  int    `bson:"newImageCount"`
	RoleName       string `bson:"roleName"`
	UserImageCount int    `bson:"userImageCount"`
	WantRole       string `bson:"wantRole"`

	Affection  int       `bson:"affection"`
	Age        int       `bson:"age"`
	Height     int       `bson:"height"`
	Weight     int       `bson:"weight"`
	Ratio      int       `bson:"ratio"`
	CreateTime time.Time `bson:"create_time"`
	Horoscope  int       `bson:"horoscope"`
}

func (this *ActiveUserLocation) TableName() string {
	return "active_user_location"
}

func (this *ActiveUserLocationModule) QueryOneByUserId(userId int64) (ActiveUserLocation, error) {
	var aul ActiveUserLocation
	c := this.session.DB("rela_match").C(aul.TableName())
	err := c.Find(bson.M{
		"userId": userId,
	}).One(&aul)
	return aul, err
}

func (this *ActiveUserLocationModule) QueryByUserIds(userIds []int64) ([]ActiveUserLocation, error) {
	var aul ActiveUserLocation
	auls := make([]ActiveUserLocation, 0)
	redisPool := factory.CacheCluster.GetConn().(*redis2.Cluster)
	rds := redisPool.NewBatch()
	var userStrs = make([]interface{}, 0)
	for _, id := range userIds {
		rds.Put("GET", "app_user_location:"+utils.GetString(id))
	}
	reply, err := redisPool.RunBatch(rds)
	if err != nil {
		log.Error(err.Error())
	}

	_, err = redis2.Scan(reply, userStrs...)
	if err != nil {
		log.Error(err)
	}

	users := make([]ActiveUserLocation, 0)
	var findUserIds = make([]int64, 0)
	for _, str := range userStrs {
		var user ActiveUserLocation
		if err := json.Unmarshal(([]byte)(utils.GetString(str)), &user); err != nil {
			log.Error(err.Error())
		} else {
			users = append(users, user)
			findUserIds = append(findUserIds, user.UserId)
		}
	}

	var notFoundUserIds = make([]int64, 0)
	for _, uId := range userIds {
		var found = false
		for _, fUid := range findUserIds {
			if fUid == uId {
				found = true
				break
			}
		}
		if !found {
			notFoundUserIds = append(notFoundUserIds, uId)
		}
	}

	c := this.session.DB("rela_match").C(aul.TableName())
	err = c.Find(bson.M{
		"userId": bson.M{
			"$in": notFoundUserIds,
		},
	}).All(&auls)

	rds = redisPool.NewBatch()
	for _, aul := range auls {
		if str, err := json.Marshal(&aul); err == nil {
			log.Infof("SET KEY: %s", "app_user_location:"+utils.GetString(aul.UserId))
			rds.Put("SETEX", "app_user_location:"+utils.GetString(aul.UserId), 600, str)
		} else {
			log.Error(err.Error())
		}
	}
	_, err = redisPool.RunBatch(rds)
	if err != nil {
		log.Error(err.Error())
	}

	var ret = make([]ActiveUserLocation, 0)
	for _, v1 := range users {
		ret = append(ret, v1)
	}
	for _, v2 := range auls {
		ret = append(ret, v2)
	}

	return ret, err
}

func (this *ActiveUserLocationModule) QueryByUserAndUsers(userId int64, userIds []int64) (ActiveUserLocation, []ActiveUserLocation, error) {
	allIds := append(userIds, userId)
	users, err := this.QueryByUserIds(allIds)
	var resUser ActiveUserLocation
	var resUsers []ActiveUserLocation
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

func (this *ActiveUserLocationModule) QueryNeighbors(lng, lat float64, notIn []int64, limit int) ([]ActiveUserLocation, error) {
	var aul ActiveUserLocation
	var auls = make([]ActiveUserLocation, 0)
	c := this.session.DB("rela_match").C(aul.TableName())
	err := c.Find(bson.M{
		"loc": bson.M{
			"$nearSphere": bson.M{
				"$geometry": bson.M{
					"type":        "Point",
					"coordinates": []float64{lng, lat},
				},
				"$maxDistance": 500 * 1000, //单位米
			},
		},
		"userId": bson.M{
			"$not": bson.M{
				"$in": notIn,
			},
		},
	}).Limit(limit).All(&auls)
	return auls, err
}

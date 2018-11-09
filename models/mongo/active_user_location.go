package mongo

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
	"errors"
	"fmt"
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
	c := this.session.DB("rela_match").C(aul.TableName())
	err := c.Find(bson.M{
		"userId": bson.M{
			"$in": userIds,
		},
	}).All(&auls)
	return auls, err
}

func (this *ActiveUserLocationModule) QueryByUserAndUsers(userId int64, userIds []int64) (ActiveUserLocation, []ActiveUserLocation, error) {
	allIds := append(userIds, userId)
	users, err := this.QueryByUserIds(allIds)
	var resUser ActiveUserLocation
	var resUsers []ActiveUserLocation
	if err == nil {
		j := 0
		for i, user := range users {
			if user.UserId == userId {
				resUser = user
				resUsers = append(users[:i], users[i+1:]...)
				ii, jj := i, j
				id1, id2 := users[ii].UserId, users[jj].UserId
				fmt.Print("findUser", i, userId, resUser.UserId, users[i].UserId, users[j].UserId, users[ii].UserId, users[jj].UserId, id1, id2)
				break
			}
			j++
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

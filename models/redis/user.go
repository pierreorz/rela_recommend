package redis

import (
	"errors"
	"strings"

	// "encoding/json"
	// "rela_recommend/log"
	"rela_recommend/algo"
	"rela_recommend/cache"
	"rela_recommend/utils"
)

type Location struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type UserProfile struct {
	UserId         int64    `json:"id"`             // 用户ID
	Location       Location `json:"location"`       //地理位置
	Avatar         string   `json:"avatar"`         // 头像
	IsVip          int      `json:"isVip"`          // 是否是vip
	LastUpdateTime int64    `json:"lastUpdateTime"` //最后在线时间
	MomentsCount   int      `json:"momentsCount"`   // 日志数
	NewImageCount  int      `json:"newImageCount"`
	RoleName       string   `json:"roleName"`
	UserImageCount int      `json:"userImageCount"`
	WantRole       string   `json:"wantRole"`
	Status         int      `json:"status"`
	Affection      int      `json:"affection"`
	Age            int      `json:"age"`
	Height         int      `json:"height"`
	Weight         int      `json:"weight"`
	Ratio          int      `json:"ratio"`
	CreateTime     JsonTime `json:"createTime"`
	Horoscope      string   `json:"horoscope"`

	JsonRoleLike map[string]float32 `json:"jsonRoleLike"`
	JsonAffeLike map[string]float32 `json:"jsonAffeLike"`
}

func (self *UserProfile) GetRoleNameInt() int {
	return utils.GetInt(self.RoleName)
}

func (self *UserProfile) GetWantRoleInts() []int {
	var wantRoles []string
	if strings.Contains(self.WantRole, ",") {
		wantRoles = strings.Split(self.WantRole, ",")
	} else {
		wantRoles = strings.Split(self.WantRole, "")
	}
	return utils.GetInts(wantRoles)
}

type UserCacheModule struct {
	CachePikaModule
}

func NewUserCacheModule(ctx algo.IContext, cache *cache.Cache, store *cache.Cache) *UserCacheModule {
	return &UserCacheModule{CachePikaModule{ctx: ctx, cache: *cache, store: *store}}
}

func (self *UserCacheModule) QueryUserById(id int64) (*UserProfile, error) {
	ids := []int64{id}
	if users, err := self.QueryUsersByIds(ids); err == nil && len(users) > 0 {
		return &users[0], nil
	}
	return nil, errors.New("not found user")
}

// 读取用户信息
func (self *UserCacheModule) QueryUsersByIds(ids []int64) ([]UserProfile, error) {
	keyFormatter := "app_user_location:%d"
	ress, err := self.MGetStructs(UserProfile{}, ids, keyFormatter, 24*60*60, 60*60*1)
	objs := ress.Interface().([]UserProfile)
	return objs, err
}

// 获取当前用户和用户列表
func (this *UserCacheModule) QueryByUserAndUsers(userId int64, userIds []int64) (UserProfile, []UserProfile, error) {
	allIds := append(userIds, userId)
	users, err := this.QueryUsersByIds(allIds)
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
			err = errors.New("user is nil" + utils.GetString(userId))
		}
	}
	return resUser, resUsers, err
}

func (this *UserCacheModule) QueryUsersMap(userIds []int64) (map[int64]*UserProfile, error) {
	users, err := this.QueryUsersByIds(userIds)
	usersMap := make(map[int64]*UserProfile, 0)
	for i, u := range users {
		if u.UserId > 0 {
			usersMap[u.UserId] = &users[i]
		}
	}
	return usersMap, err
}

func (this *UserCacheModule) QueryByUserAndUsersMap(userId int64, userIds []int64) (*UserProfile, map[int64]*UserProfile, error) {
	user, users, err := this.QueryByUserAndUsers(userId, userIds)
	usersMap := make(map[int64]*UserProfile, 0)
	for i, u := range users {
		if u.UserId > 0 {
			usersMap[u.UserId] = &users[i]
		}
	}
	return &user, usersMap, err
}

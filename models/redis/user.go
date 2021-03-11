package redis

import (
	"errors"
	"fmt"
	"strings"

	// "encoding/json"
	// "rela_recommend/log"
	"rela_recommend/cache"
	"rela_recommend/service/abtest"
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
	Reason         string   `json:"reason"` //优质用户推荐理由
	Grade          float64  `json:"grade"`  //优质用户推荐等级 1-100

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

func NewUserCacheModule(ctx abtest.IAbTestAble, cache *cache.Cache, store *cache.Cache) *UserCacheModule {
	return &UserCacheModule{CachePikaModule{ctx: ctx, cache: *cache, store: *store}}
}

func (self *UserCacheModule) QueryUserById(id int64) (*UserProfile, error) {
	ids := []int64{id}
	if users, err := self.QueryUsersByIds(ids); err == nil && len(users) > 0 {
		return &users[0], nil
	}
	return nil, errors.New(fmt.Sprintf("not found user[%d]", id))
}

// 读取用户信息
func (self *UserCacheModule) QueryUsersByIds(ids []int64) ([]UserProfile, error) {
	keyFormatter := self.ctx.GetAbTest().GetString("user_cache_key_formatter", "app_user_active_info_search_%d")
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
	// log.Infof("QueryByUserAndUsers: users:%+v\n", users)
	if err == nil {
		for i, user := range users {
			if user.UserId == userId {
				resUser = user
				resUsers = append(users[:i], users[i+1:]...)
				// users i后面的内容向前移动了一位，内容发上了改变，谨慎使用
				break
			}
		}
		if resUser.UserId == 0 { // 如果找不到用户，则返回其他列表
			resUsers = users
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
	// log.Infof("QueryByUserAndUsersMap: user:%+v users:%+v\n", user, users)
	for i, u := range users {
		if u.UserId > 0 {
			usersMap[u.UserId] = &users[i]
		}
	}
	return &user, usersMap, err
}

// 查询用户关注列表，依赖缓冲，后期使用接口替换
func (this *UserCacheModule) QueryConcernsByUser(userId int64) ([]int64, error) {
	return this.SmembersInt64List(userId, "user_concern:%d")
}

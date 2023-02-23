package redis

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

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

type liveInfo struct {
	Status     int   `json:"status"`
	ExpireDate int64 `json:"expire_date"`
}

type UserProfile struct {
	UserId         int64              `json:"id"`             // 用户ID
	Location       Location           `json:"location"`       //地理位置
	Avatar         string             `json:"avatar"`         // 头像
	IsVip          int                `json:"isVip"`          // 是否是vip
	LastUpdateTime int64              `json:"lastUpdateTime"` //最后在线时间
	MomentsCount   int                `json:"momentsCount"`   // 日志数
	NewImageCount  int                `json:"newImageCount"`
	RoleName       string             `json:"roleName"`
	UserImageCount int                `json:"userImageCount"`
	WantRole       string             `json:"wantRole"`
	Status         int                `json:"status"`
	Affection      int                `json:"affection"`
	Age            int                `json:"age"`
	Height         int                `json:"height"`
	Weight         int                `json:"weight"`
	Ratio          int                `json:"ratio"`
	CreateTime     JsonTime           `json:"createTime"`
	Horoscope      string             `json:"horoscope"`
	Reason         string             `json:"reason"` //优质用户推荐理由
	Grade          float64            `json:"grade"`  //优质用户推荐等级 1-100
	Recall         int                `json:"new_recall,omitempty"`
	ActiveDate     string             `json:"active_date"`           // 用于计算回流用户
	LastActiveDate string             `json:"last_active_date"`      // 用于计算回流用户
	Intro          string             `json:"intro"`                 //用户标签
	Occupation     string             `json:"occupation"`            //用户职业
	TeenActive     int8               `json:"teen_active,omitempty"` //是否是青少年模式
	IsPrivate      int                `json:"is_private,omitempty"`
	JsonRoleLike   map[string]float32 `json:"jsonRoleLike"`
	JsonAffeLike   map[string]float32 `json:"jsonAffeLike"`
	LiveInfo       *liveInfo          `json:"live_info,omitempty"`
	OnlineHiding   int8               `json:"online_hiding,omitempty"`
	Hiding         int8               `json:"hiding,omitempty"`
	Identity       int8               `json:"identity"` //身份认证 -1 未通过，0 未验证，1 审核中，2系统认证通过，3 人工认证通过，4 主播认证通过，5  历史申诉认证通过
}

type UserContentProfile struct {
	UserId      int64              `json:"user_id"`
	PicturePref map[string]float32 `json:"picture_pref,omitempty"`
}

type UserLiveContentProfile struct {
	UserId       int64             `json:"user_id"`
	WantRole     int               `json:"want_role"`
	UserLivePref map[int64]float64 `json:"user_live_pref,omitempty"`
}

type LiveContentProfile struct {
	LiveId           int64   `json:"live_id"`
	LiveContentScore float64 `json:"live_content_score"`
	LiveValueScore   float64 `json:"live_value_score"`
	Role             int     `json:"role_type"`
}

type UserLiveProfile struct {
	UserId                int64             `json:"user_id"`                      //用户id
	LiveLongPref          map[int64]float32 `json:"live_long_pref,omitempty"`     //用户长期主播偏好
	LiveShortPref         map[int64]float32 `json:"live_short_pref,omitempty"`    //用户短期主播偏好
	ConsumeLongPref       map[int64]float32 `json:"consume_long_pref,omitempty"`  //用户长期消费偏好
	ConsumeShortPref      map[int64]float32 `json:"consume_short_pref,omitempty"` //用户短期消费偏好
	LiveTypeLongPref      map[int]float32   `json:"live_type_long_pref,omitempty"`
	LiveTypeShortPref     map[int]float32   `json:"live_type_short_pref,omitempty"`
	LiveClassifyLongPref  map[int]float32   `json:"live_classify_long_pref,omitempty"`
	LiveClassifyShortPref map[int]float32   `json:"live_classify_short_pref,omitempty"`
}

//读取用户直播画像
func (self *UserCacheModule) QueryUserLiveProfileByIds(ids []int64) ([]UserLiveProfile, error) {
	keyFormatter := "user_live_profile:%d"
	ress, err := self.MGetStructs(UserLiveProfile{}, ids, keyFormatter, 24*60*60, 60*60*1)
	objs := ress.Interface().([]UserLiveProfile)
	return objs, err
}

// 获取当前用户和用户列表Map
func (this *UserCacheModule) QueryUserLiveProfileByIdsMap(userIds []int64) (map[int64]*UserLiveProfile, error) {
	userLiveProfiles, err := this.QueryUserLiveProfileByIds(userIds)
	var resUserLiveProfileMap = make(map[int64]*UserLiveProfile, 0)
	if err == nil {
		for i, user := range userLiveProfiles {
			resUserLiveProfileMap[user.UserId] = &userLiveProfiles[i]
		}
	}
	return resUserLiveProfileMap, err
}

//读取用户内容画像
func (self *UserCacheModule) QueryUserContentProfileByIds(ids []int64) ([]UserContentProfile, error) {
	keyFormatter := "user_picture_pref:%d"
	ress, err := self.MGetStructs(UserContentProfile{}, ids, keyFormatter, 24*60*60, 60*60*1)
	objs := ress.Interface().([]UserContentProfile)
	return objs, err
}

// 获取当前用户和用户列表Map
func (this *UserCacheModule) QueryUserContentProfileByIdsMap(userIds []int64) (map[int64]*UserContentProfile, error) {
	userContentProfiles, err := this.QueryUserContentProfileByIds(userIds)
	var resUserContentProfileMap = make(map[int64]*UserContentProfile, 0)
	if err == nil {
		for i, user := range userContentProfiles {
			resUserContentProfileMap[user.UserId] = &userContentProfiles[i]
		}
	}
	return resUserContentProfileMap, err
}

//读取主播画像数据

func (self *UserCacheModule) QueryLiveContentProfileByIds(ids []int64) ([]LiveContentProfile, error) {
	keyFormatter := "live_content_profile:%d"
	ress, err := self.MGetStructs(LiveContentProfile{}, ids, keyFormatter, 24*60*60, 60*60*1)
	objs := ress.Interface().([]LiveContentProfile)
	return objs, err
}

// 获取当前用户和用户列表Map
func (this *UserCacheModule) QueryLiveContentProfileByIdsMap(userIds []int64) (map[int64]*LiveContentProfile, error) {
	liveContentProfiles, err := this.QueryLiveContentProfileByIds(userIds)
	var resUserContentProfileMap = make(map[int64]*LiveContentProfile, 0)
	if err == nil {
		for i, user := range liveContentProfiles {
			resUserContentProfileMap[user.LiveId] = &liveContentProfiles[i]
		}
	}
	return resUserContentProfileMap, err
}

//读取用户画像数据
func (self *UserCacheModule) QueryUserLiveContentProfileByIds(ids []int64) ([]UserLiveContentProfile, error) {
	keyFormatter := "user_live_content_profile:%d"
	ress, err := self.MGetStructs(UserLiveContentProfile{}, ids, keyFormatter, 24*60*60, 60*60*1)
	objs := ress.Interface().([]UserLiveContentProfile)
	return objs, err
}

// 获取当前用户和用户列表Map
func (this *UserCacheModule) QueryUserLiveContentProfileByIdsMap(userIds []int64) (map[int64]*UserLiveContentProfile, error) {
	liveContentProfiles, err := this.QueryUserLiveContentProfileByIds(userIds)
	var resUserContentProfileMap = make(map[int64]*UserLiveContentProfile, 0)
	if err == nil {
		for i, user := range liveContentProfiles {
			resUserContentProfileMap[user.UserId] = &liveContentProfiles[i]
		}
	}
	return resUserContentProfileMap, err
}


func (user *UserProfile) InChina() bool {
	if user.Location.Lat>=4&&user.Location.Lat<=53&&
		user.Location.Lon>=74&&user.Location.Lon<=135{
		return true
	}
	return false
}
func (user *UserProfile) MaybeICPUser(lat, lng float32) bool {
	// 特定ICP审核用户
	if user.UserId == 104208008 {
		return true
	}

	// 杭州经纬度的新注册用户(大于 2021-03-01 00:00:00)
	// 也可能没打开经纬度
	if user.CreateTime.Unix() > 1614528000 {
		//if user.CreateTime.Unix() > 1623945600 {
		//	if user.Location.Lat >= 30.600137 && user.Location.Lat <= 31.336124 &&
		//		user.Location.Lon >= 121.135242 && user.Location.Lon <= 122.081039 {
		//		return true
		//	}

		if lat >= 30.043946 && lat <= 30.466238 &&
			lng >= 119.892146 && lng <= 120.595841 {
			return true
		}

		if math.Abs(user.Location.Lon-0.0) <= 1e-6 || math.Abs(user.Location.Lat-0.0) <= 1e-6 {
			return true
		}
	}
	return false
}

func (user *UserProfile) IsVipHiding() bool {
	if user == nil {
		return false
	}
	if user.IsVip == 1 && user.Hiding == 1 {
		return true
	}
	return false
}

func (user *UserProfile) IsVipHidingMom() bool {
	if user == nil {
		return false
	}
	if user.IsVip == 1 && user.Hiding == 1 {
		return true
	}
	return false
}

// 使用者测是否可以推荐
func (user *UserProfile) CanRecommend() bool {

	if user == nil {
		return false
	}

	notTeen := user.Status != 7 && user.TeenActive != 1 // 非青少年模式

	return notTeen
}

// DataUserCanRecommend 被推荐内容的用户是否可以推荐
func (user *UserProfile) DataUserCanRecommend() bool {
	if user == nil {
		return false
	}

	isNormal := user.Status == 1                                // 状态正常
	notPrivate := user.IsPrivate == 0                           // 非私密账号
	femaleIdentity := user.Identity != -1 && user.Identity != 1 // 非男性用户

	return isNormal && notPrivate && femaleIdentity
}

// DataUserCandidateCanRecommend 被推荐内容的用户是否可以推荐
func (user *UserProfile) DataUserCandidateCanRecommend() bool {
	if user == nil {
		return false
	}

	isNormal := user.Status == 1                                // 状态正常
	femaleIdentity := user.Identity != -1 && user.Identity != 1 // 非男性用户

	return isNormal && femaleIdentity
}

func (user *UserProfile) GetRoleNameInt() int {
	return utils.GetInt(user.RoleName)
}

func (user *UserProfile) IsRecurringUser(compareTime time.Time, threshold time.Duration) bool {
	// 在 compareTime 当天活跃，且距离上一次活跃超过 threshold 时间
	var activeDate, lastActiveDate time.Time

	if len(user.ActiveDate) > 0 {
		activeDate, _ = time.ParseInLocation("2006.01.02", user.ActiveDate, time.Local)
	}

	if len(user.LastActiveDate) > 0 {
		lastActiveDate, _ = time.ParseInLocation("2006.01.02", user.LastActiveDate, time.Local)
	}

	if (compareTime.Year() == activeDate.Year()) &&
		(compareTime.Month() == activeDate.Month()) &&
		(compareTime.Day() == activeDate.Day()) && (activeDate.Sub(lastActiveDate) >= threshold) {
		return true
	}

	return false
}

func (user *UserProfile) GetWantRoleInts() []int {
	var wantRoles []string
	if strings.Contains(user.WantRole, ",") {
		wantRoles = strings.Split(user.WantRole, ",")
	} else {
		wantRoles = strings.Split(user.WantRole, "")
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
	//log.Infof("QueryByUserAndUsersMap: user:%+v users:%+v\n", user, users)
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

func (this *UserCacheModule) QueryConcernsByUserV1(userId int64) ([]int64, error) {
	return this.ZmembersInt64List(userId, "user:%d:followers")
}

package redis

import "rela_recommend/log"

type MatchProfile struct {
	UserID       int64              `json:"user_id"`
	AgeMap       map[string]float32 `json:"age"`
	RoleNameMap  map[string]float32 `json:"role_name"`
	HoroscopeMap map[string]float32 `json:"horoscope"`
	HeightMap    map[string]float32 `json:"height"`
	WeightMap    map[string]float32 `json:"weight"`
	DistanceMap  map[string]float32 `json:"distance"`
	LikeTypeMap  map[string]float32 `json:"like_type"`
	AffectionMap map[string]float32 `json:"affection"`
	MobileSysMap map[string]float32 `json:"mobile_sys"`
	TotalCount   int64              `json:"total_count"`
	FreqWeekMap  map[string]float32 `json:"freq_week"`
	FreqTimeMap  map[string]float32 `json:"freq_time"`
	ContinuesUse int64              `json:"continues_use"`
	// TimestampMap  map[string]float32 `json:"timestamp"`
}


// 读取速配画像信息
func (self *UserCacheModule) QueryMatchProfileByIds(ids []int64) ([]MatchProfile, error) {
	keyFormatter := "match_user_profile:%d"
	ress, err := self.MGetStructs(MatchProfile{}, ids, keyFormatter, 24 * 60 * 60, 60 * 60 * 1)
	objs := ress.Interface().([]MatchProfile)
	log.Warnf("match redis err,%s,%s",ress,err)
	return objs, err
}

// 获取当前用户和用户列表Map
func (this *UserCacheModule) QueryMatchProfileByUserAndUsersMap(userId int64, userIds []int64) (*MatchProfile, map[int64]*MatchProfile, error) {
	allIds := append(userIds, userId)
	users, err := this.QueryMatchProfileByIds(allIds)
	var resUser *MatchProfile
	var resUsersMap = make(map[int64]*MatchProfile, 0)
	if err == nil {
		for _, user := range users {
			if user.UserID == userId {
				resUser = &user
			} else {
				resUsersMap[user.UserID] = &user
			}
		}
	}
	return resUser, resUsersMap, err
}

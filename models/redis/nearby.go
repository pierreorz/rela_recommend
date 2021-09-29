package redis

type NearbyProfile struct {
	UserID        int64                       `json:"user_id"`
	TotalMap      map[string]float32          `json:"total"`
	AgeMap        map[string]float32          `json:"age"`
	RoleNameMap   map[string]float32          `json:"role_name"`
	AffectionMap  map[string]float32          `json:"affection"`
	HoroscopeMap  map[string]float32          `json:"horoscope"`
	HeightMap     map[string]float32          `json:"height"`
	WeightMap     map[string]float32          `json:"weight"`
	DistanceMap   map[string]float32          `json:"dis"`
	MobileSysMap  map[string]float32          `json:"mobile_sys"`
	ActiveTimeMap map[string]float32          `json:"active_time"`
	FreqWeekMap   map[string]float32          `json:"week"`
	FreqTimeMap   map[string]float32          `json:"time"`
	Last30dMap    map[string]float32          `json:"last_30d"`
	Last7dMap     map[string]float32          `json:"last_7d"`
	NearSeeMap    map[string]float32          `json:"near_see"`
	NearShowMap   map[string]float32          `json:"near_show"`
	VectorMap     map[string][]float32        `json:"vector"`
	WeekExposures map[string]exposureAndClick `json:"week_exposures"`
}

type exposureAndClick struct {
	Exposures int `json:"exposures"`
	Clicks    int `json:"clicks"`
}

// 读取速配画像信息
func (self *UserCacheModule) QueryNearbyProfileByIds(ids []int64, keyFormatter string) ([]NearbyProfile, error) {
	ress, err := self.MGetStructs(NearbyProfile{}, ids, keyFormatter, 24*60*60, 60*60*1)
	objs := ress.Interface().([]NearbyProfile)
	return objs, err
}

// 获取当前用户和用户列表Map
func (this *UserCacheModule) QueryNearbyProfileByUserAndUsersMap(userId int64, userIds []int64, keyFormatter string) (*NearbyProfile, map[int64]*NearbyProfile, error) {
	allIds := append(userIds, userId)
	users, err := this.QueryNearbyProfileByIds(allIds, keyFormatter)
	var resUser *NearbyProfile
	var resUsersMap = make(map[int64]*NearbyProfile, 0)
	if err == nil {
		for i, user := range users {
			if user.UserID == userId {
				resUser = &users[i]
			} else {
				resUsersMap[user.UserID] = &users[i]
			}
		}
	}
	return resUser, resUsersMap, err
}

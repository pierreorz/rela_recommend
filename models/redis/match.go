package redis

type MatchProfile struct {
	UserID        int64              `json:"user_id"`
	AgeMap        map[string]float32 `json:"age"`
	RoleNameMap   map[string]float32 `json:"role_name"`
	HoroscopeMap  map[string]float32 `json:"horoscope"`
	HeightMap     map[string]float32 `json:"height"`
	WeightMap     map[string]float32 `json:"weight"`
	DistanceMap   map[string]float32 `json:"distance"`
	LikeTypeMap   map[string]float32 `json:"like_type"`
	AffectionMap  map[string]float32 `json:"affection"`
	MobileSysMap  map[string]float32 `json:"mobile_sys"`
	TotalCountMap map[string]float32 `json:"total_count"`
	FreqWeekMap   map[string]float32 `json:"freq_week"`
	FreqTimeMap   map[string]float32 `json:"freq_time"`
	ContinuesUse  int64              `json:"continues_use"`
	// TimestampMap  map[string]float32 `json:"timestamp"`
}

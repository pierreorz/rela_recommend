package quick_match

import (
	"rela_recommend/algo"
	"rela_recommend/models/mongo"
)

type UserInfo struct {
	UserId int64
	UserCache *mongo.ActiveUserLocation
	Score float32
	Features []algo.Feature
}

type QuickMatchContext struct {
	User *UserInfo
	UserList []UserInfo
}

// 用户排序
type UserInfoSortReverse []UserInfo

func (a UserInfoSortReverse) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a UserInfoSortReverse) Len() int {
	return len(a)
}

// 以此按照：打分，资料完成度，
func (a UserInfoSortReverse) Less(i, j int) bool {
	if a[i].Score == a[j].Score {
		return a[i].UserCache.LastUpdateTime < a[j].UserCache.LastUpdateTime
	}
	return a[i].Score < a[j].Score
}

package theme

import(
	"time"
	"rela_recommend/algo"
	rutils "rela_recommend/utils"
	"rela_recommend/models/pika"
	"rela_recommend/algo/utils"
	"rela_recommend/service/abtest"
)

type RankInfo struct {
	IsTop		int 		// 1: 置顶， 0: 默认， -1:置底
	Level		int			// 推荐等级
	Punish		float32		// 惩罚系数
	AlgoScore 	float32		// 算法得分
	Score 		float32		// 最终得分
}

// 用户信息
type UserInfo struct {
	UserId int64
	UserCache *pika.UserProfile
	UserConcerns *rutils.SetInt64
}

// 话题信息
type ThemeInfo struct {
	UserId 		int64
	UserCache 	*pika.UserProfile
	RankInfo	*RankInfo
	Features 	*utils.Features
}

// 直播推荐算法上下文
type ThemeAlgoContext struct {
	RankId string
	CreateTime time.Time
	Platform int
	Request *algo.RecommendRequest
	AbTest *abtest.AbTest
	User *UserInfo
	ThemeIds []int64
	ThemeList []ThemeInfo
}

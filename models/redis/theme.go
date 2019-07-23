package redis

import(
	"fmt"
	// "encoding/json"
	// "rela_recommend/log"
	"rela_recommend/cache"
	"rela_recommend/algo"
)

type Behavior struct {
	Count			float64		`json:"count"`
	LastTime		float64		`json:"last_time"`
}

// 话题用户行为缓存
type ThemeUserBehavior struct {
	CacheTime      			float64 	`json:"cache_time,omitempty"`				// 缓存时间
	LastTime				float64 	`json:"last_time,omitempty"`				// 最后动作时间

	ListExposure			*Behavior 	`json:"theme.list:exposure,omitempty"`		// 列表曝光
	ListClick				*Behavior 	`json:"theme.list:click,omitempty"`			// 列表曝光

	DetailLike				*Behavior 	`json:"theme.detail:like"`					// 详情页喜欢
	DetailUnLike			*Behavior 	`json:"theme.detail:unlike"`				// 详情页喜欢
	DetailExposure			*Behavior	`json:"theme.detail:exposure"`				// 详情页曝光
	DetailComment			*Behavior	`json:"theme.detail:comment"`				// 详情页评论
	DetailShare				*Behavior	`json:"theme.detail:share"`					// 详情页分享
	DetailFollowThemer		*Behavior	`json:"theme.detail:follow_themer"`			// 详情页关注
	DetailUnFollowThemer	*Behavior	`json:"theme.detail:unfollow_themer"`		// 详情页关注

	DetailLikeReply			*Behavior 	`json:"theme.detail:like_reply"`			// 详情页评论喜欢
	DetailUnLikeReply		*Behavior 	`json:"theme.detail:unlike_reply"`			// 详情页评论喜欢
	DetailExposureReply		*Behavior	`json:"theme.detail:exposure_reply"`		// 详情页评论曝光
	DetailCommentReply		*Behavior 	`json:"theme.detail:comment_reply"`			// 详情页评论评论
	DetailShareReply		*Behavior	`json:"theme.detail:share_reply"`			// 详情页评论分享
	DetailFollowReply		*Behavior	`json:"theme.detail:follow_replyer"`		// 关注评论者
	DetailUnFollowReply		*Behavior	`json:"theme.detail:unfollow_replyer"`		// 关注评论者
}

// 是否曝光
func (self *ThemeUserBehavior) IsListExposured() bool {
	return self.ListExposure != nil && self.ListExposure.Count > 0
}

// 点击率
func (self *ThemeUserBehavior) ListClickRate() float64 {
	if self.IsListExposured() {
		if self.ListClick != nil {
			return self.ListClick.Count / self.ListExposure.Count
		}
	}
	return 0.0
}


type ThemeBehaviorCacheModule struct {
	CachePikaModule
}

func NewThemeBehaviorCacheModule(ctx algo.IContext, cache *cache.Cache) *ThemeBehaviorCacheModule {
	return &ThemeBehaviorCacheModule{CachePikaModule{ctx: ctx, cache: *cache, store: nil}}
}

// 读取话题相关用户行为
func (self *ThemeBehaviorCacheModule) QueryUserBehaviorMap(userId int64, ids []int64) (map[int64]*ThemeUserBehavior, error) {
	keyFormatter := fmt.Sprintf("behavior:theme:%d:%%d", userId)
	ress, err := self.MGetStructsMap(&ThemeUserBehavior{}, ids, keyFormatter, 0, 0)
	objs := ress.Interface().(map[int64]*ThemeUserBehavior)
	return objs, err
}

// 读取话题相关行为
func (self *ThemeBehaviorCacheModule) QueryBehaviorMap(ids []int64) (map[int64]*ThemeUserBehavior, error) {
	keyFormatter := "behavior:theme:%d"
	ress, err := self.MGetStructsMap(&ThemeUserBehavior{}, ids, keyFormatter, 0, 0)
	objs := ress.Interface().(map[int64]*ThemeUserBehavior)
	return objs, err
}
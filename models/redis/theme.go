package redis

import(
	"fmt"
	// "encoding/json"
	// "rela_recommend/log"
	"rela_recommend/cache"
	"rela_recommend/algo"
)

// 话题用户行为缓存
type ThemeUserBehavior struct {
	CacheTime      			float64 	`json:"cache_time,omitempty"`					// 缓存时间
	LastTime				float64 	`json:"last_time,omitempty"`					// 最后动作时间
	Count					float64		`json:"count,omitempty"`						// 触发动作次数

	ListExposure			*Behavior 	`json:"theme.list:exposure,omitempty"`			// 列表曝光
	ListClick				*Behavior 	`json:"theme.list:click,omitempty"`				// 列表曝光
	ListRecommendExposure	*Behavior 	`json:"theme.recommend:exposure,omitempty"`		// 列表曝光
	ListRecommendClick		*Behavior 	`json:"theme.recommend:click,omitempty"`		// 列表曝光

	DetailExposure			*Behavior 	`json:"theme.detail:exposure"`					// 详情页曝光
	DetailLike				*Behavior 	`json:"theme.detail:like_theme"`				// 详情页喜欢
	DetailUnLike			*Behavior 	`json:"theme.detail:unlike_theme"`				// 详情页喜欢
	DetailComment			*Behavior	`json:"theme.detail:comment"`					// 详情页评论
	DetailShare				*Behavior	`json:"theme.detail:share_theme"`				// 详情页分享
	DetailFollowThemer		*Behavior	`json:"theme.detail:follow_themer"`				// 详情页关注
	DetailUnFollowThemer	*Behavior	`json:"theme.detail:unfollow_themer"`			// 详情页关注

	DetailExposureReply		*Behavior	`json:"theme.detail:exposure_reply"`			// 详情页评论曝光
	DetailLikeReply			*Behavior 	`json:"theme.detail:like_reply"`				// 详情页评论喜欢
	DetailUnLikeReply		*Behavior 	`json:"theme.detail:unlike_reply"`				// 详情页评论喜欢
	DetailCommentReply		*Behavior 	`json:"theme.detail:comment_reply"`				// 详情页评论评论
	DetailShareReply		*Behavior	`json:"theme.detail:share_reply"`				// 详情页评论分享
	DetailFollowReplyer		*Behavior	`json:"theme.detail:follow_replyer"`			// 关注评论者
	DetailUnFollowReplyer	*Behavior	`json:"theme.detail:unfollow_replyer"`			// 取消关注评论者
}
type ThemeUserProfile struct {
	UserID 	int64	`json:"user_id"`
	UserEmbedding []float32 `json:"user_embedding"`

}
type ThemeProfile struct {
	ThemeID int64	`json:"theme_id"`
	ThemeEmbedding []float32 `json:"theme_embedding"`
}

// 获取总列表曝光
func (self *ThemeUserBehavior) GetTotalListExposure() *Behavior {
	return MergeBehaviors(self.ListExposure, self.ListRecommendExposure)
}

// 获取总列表曝光
func (self *ThemeUserBehavior) GetTotalListClick() *Behavior {
	return MergeBehaviors(self.ListClick, self.ListRecommendClick)
}

// 获取总交互汇总
func (self *ThemeUserBehavior) GetTotalInteract() *Behavior {
	return MergeBehaviors(
		self.DetailLike, self.DetailLikeReply, 
		self.DetailComment, self.DetailCommentReply,
		self.DetailShare, self.DetailShareReply,
		self.DetailFollowThemer, self.DetailFollowReplyer,)
}


type ThemeBehaviorCacheModule struct {
	CachePikaModule
}

func NewThemeBehaviorCacheModule(ctx algo.IContext, cache *cache.Cache) *ThemeBehaviorCacheModule {
	return &ThemeBehaviorCacheModule{CachePikaModule{ctx: ctx, cache: *cache, store: nil}}
}

type ThemeUserProfileModule struct {
	CachePikaModule
}

func NewThemeCacheModule(ctx algo.IContext, cache *cache.Cache, store *cache.Cache) *ThemeUserProfileModule{
	return &ThemeUserProfileModule{CachePikaModule{ctx: ctx, cache: *cache, store: *store}}
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

// 读取用户als特征
func (self *ThemeUserProfileModule) QueryThemeUserProfileMap(ids []int64) (map[int64]*ThemeUserProfile, error) {
	keyFormatter := "theme_user_profile:%d"
	ress, err := self.MGetStructsMap(&ThemeUserProfile{}, ids, keyFormatter, 0, 0)
	objs := ress.Interface().(map[int64]*ThemeUserProfile)
	return objs, err
}
func (self *ThemeUserProfileModule) QueryThemeProfileMap(ids []int64) (map[int64]*ThemeProfile, error) {
	keyFormatter := "theme_profile:%d"
	ress, err := self.MGetStructsMap(&ThemeProfile{}, ids, keyFormatter, 0, 0)
	objs := ress.Interface().(map[int64]*ThemeProfile)
	return objs, err
}




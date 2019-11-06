package behavior

// ListExposure			*Behavior 	`json:"theme.list:exposure,omitempty"`			// 列表曝光
// ListClick				*Behavior 	`json:"theme.list:click,omitempty"`				// 列表曝光
// ListRecommendExposure	*Behavior 	`json:"theme.recommend:exposure,omitempty"`		// 列表曝光
// ListRecommendClick		*Behavior 	`json:"theme.recommend:click,omitempty"`		// 列表曝光

// DetailExposure			*Behavior 	`json:"theme.detail:exposure"`					// 详情页曝光
// DetailLike				*Behavior 	`json:"theme.detail:like_theme"`				// 详情页喜欢
// DetailUnLike			*Behavior 	`json:"theme.detail:unlike_theme"`				// 详情页喜欢
// DetailComment			*Behavior	`json:"theme.detail:comment"`					// 详情页评论
// DetailShare				*Behavior	`json:"theme.detail:share_theme"`				// 详情页分享
// DetailFollowThemer		*Behavior	`json:"theme.detail:follow_themer"`				// 详情页关注
// DetailUnFollowThemer	*Behavior	`json:"theme.detail:unfollow_themer"`			// 详情页关注

// DetailExposureReply		*Behavior	`json:"theme.detail:exposure_reply"`			// 详情页评论曝光
// DetailLikeReply			*Behavior 	`json:"theme.detail:like_reply"`				// 详情页评论喜欢
// DetailUnLikeReply		*Behavior 	`json:"theme.detail:unlike_reply"`				// 详情页评论喜欢
// DetailCommentReply		*Behavior 	`json:"theme.detail:comment_reply"`				// 详情页评论评论
// DetailShareReply		*Behavior	`json:"theme.detail:share_reply"`				// 详情页评论分享
// DetailFollowReplyer		*Behavior	`json:"theme.detail:follow_replyer"`			// 关注评论者
// DetailUnFollowReplyer	*Behavior	`json:"theme.detail:unfollow_replyer"`			// 取消关注评论者


// 获取列表曝光行为
func (self *UserBehavior) GetThemeListExposure() *Behavior {
	return self.Gets("theme.recommend:exposure", "theme.list:exposure", "theme.news:exposure", "theme.hotweek:exposure")
}

// 获取列表互动行为
func (self *UserBehavior) GetThemeListInteract() *Behavior {
	return self.Gets("theme.recommend:click", "theme.list:click", "theme.news:click", "theme.hotweek:click")
}

// 详情页曝光行为
func (self *UserBehavior) GetThemeDetailExposure() *Behavior {
	return self.Gets("theme.detail:exposure")
}

// 详情页互动行为
func (self *UserBehavior) GetThemeDetailInteract() *Behavior {
	return self.Gets(
		"theme.detail:like_theme", "theme.detail:comment", "theme.detail:share_theme", "theme.detail:follow_themer",
		"theme.detail:like_reply", "theme.detail:comment_reply", "theme.detail:share_reply", "theme.detail:follow_replyer",
	)
}

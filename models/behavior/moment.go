package behavior

// // 朋友圈
// FriendExposure			*Behavior 	`json:"moment.friend:exposure"`					// 朋友圈曝光
// FriendClick				*Behavior 	`json:"moment.friend:click"`					// 朋友圈点击
// FriendLike				*Behavior 	`json:"moment.friend:like"`						// 朋友圈喜欢
// FriendUnLike			*Behavior 	`json:"moment.friend:unlike"`					// 朋友圈喜欢
// FriendComment			*Behavior	`json:"moment.friend:comment"`					// 朋友圈评论
// FriendShare				*Behavior	`json:"moment.friend:share"`					// 朋友圈分享
// FriendFollow			*Behavior	`json:"moment.friend:follow"`					// 朋友圈关注
// FriendUnFollow			*Behavior	`json:"moment.friend:unfollow"`					// 朋友圈关注
// // 附近的日志
// AroundExposure			*Behavior 	`json:"aroundmoment:exposure"`					// 附近日志曝光
// AroundClick				*Behavior 	`json:"aroundmoment:click"`					// 附近日志点击
// AroundLike				*Behavior 	`json:"aroundmoment:like"`						// 附近日志喜欢
// AroundUnLike			*Behavior 	`json:"aroundmoment:unlike"`					// 附近日志喜欢
// AroundComment			*Behavior	`json:"aroundmoment:comment"`					// 附近日志评论
// AroundShare				*Behavior	`json:"aroundmoment:share"`					// 附近日志分享
// AroundFollow			*Behavior	`json:"aroundmoment:follow"`					// 附近日志关注
// AroundUnFollow			*Behavior	`json:"aroundmoment:unfollow"`					// 附近日志关注
// // 推荐日志
// RecommendExposure		*Behavior 	`json:"moment.recommend:exposure"`				// 推荐曝光
// RecommendClick			*Behavior 	`json:"moment.recommend:click"`					// 推荐列表点击
// RecommendLike			*Behavior 	`json:"moment.recommend:like"`					// 推荐喜欢
// RecommendUnLike			*Behavior 	`json:"moment.recommend:unlike"`				// 推荐喜欢
// RecommendComment		*Behavior	`json:"moment.recommend:comment"`				// 推荐评论
// RecommendShare			*Behavior	`json:"moment.recommend:share"`					// 推荐分享
// RecommendFollow			*Behavior	`json:"moment.recommend:follow"`				// 推荐关注
// RecommendUnFollow		*Behavior	`json:"moment.recommend:unfollow"`				// 推荐关注
// // 日志详情
// DetailExposure			*Behavior 	`json:"moment.detail:exposure"`					// 日志详情曝光
// DetailLike				*Behavior 	`json:"moment.detail:like"`						// 日志详情喜欢
// DetailUnLike			*Behavior 	`json:"moment.detail:unlike"`					// 日志详情喜欢
// DetailComment			*Behavior	`json:"moment.detail:comment"`					// 日志详情评论
// DetailShare				*Behavior	`json:"moment.detail:share"`					// 日志详情分享
// DetailFollow			*Behavior	`json:"moment.detail:follow"`					// 日志详情关注
// DetailUnFollow			*Behavior	`json:"moment.detail:unfollow"`					// 日志详情关注

// 获取总列表曝光
func (self *UserBehavior) GetMomentListExposure() *Behavior {
	return self.Gets("moment.friend:exposure", "moment.around:exposure", "moment.recommend:exposure")
}

// 获取总互动行为
func (self *UserBehavior) GetMomentListInteract() *Behavior {
	return self.Gets(
		"moment.friend:like", "moment.friend:comment", "moment.friend:share", "moment.friend:follow",
		"moment.around:like", "moment.around:comment", "moment.around:share", "moment.around:follow",
		"moment.recommend:like", "moment.recommend:comment", "moment.recommend:share", "moment.recommend:follow",
		"moment.detail:like", "moment.detail:comment", "moment.detail:share", "moment.detail:follow",
	)
}

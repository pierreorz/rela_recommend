package live

import (
	"math"
	"math/rand"
	"rela_recommend/algo"
	"rela_recommend/algo/utils"
	rutils "rela_recommend/utils"
)

// 处理业务给出的置顶和推荐内容
func LiveTopRecommandStrategyFunc(ctx algo.IContext, index int) error {
	dataInfo := ctx.GetDataByIndex(index)
	rankInfo := dataInfo.GetRankInfo()
	live := dataInfo.(*LiveInfo)

	if live.LiveCache.Recommand == 1 { // 1: 推荐
		if live.LiveCache.RecommandLevel > 10 { // 15: 置顶
			rankInfo.IsTop = 1
		}
		rankInfo.Level = rankInfo.Level + live.LiveCache.RecommandLevel // 推荐等级
	} else if live.LiveCache.Recommand == -1 { // -1: 不推荐
		if live.LiveCache.RecommandLevel == -1 { // -1: 置底
			rankInfo.IsTop = -1
		} else if live.LiveCache.RecommandLevel > 0 { // 降低权重
			level := math.Min(float64(live.LiveCache.RecommandLevel), 100.0)
			// rankInfo.Punish = float32(100.0 - level) / 100.0
			rankInfo.AddRecommend("Down", float32(100.0-level)/100.0)
		}
	}
	return nil
}

func LiveExposureFunc(ctx algo.IContext) error {
	abtest := ctx.GetAbTest()
	userInfo := ctx.GetUserInfo().(*UserInfo)
	videoList :=make(map[int64]int,0)
	count :=0
	if userInfo.ConsumeUser==0{//低消费用户将视频提权
		for index := 0; index < ctx.GetDataLength(); index++ {
			dataInfo := ctx.GetDataByIndex(index).(*LiveInfo)
			rankInfo := dataInfo.GetRankInfo()
			if dataInfo.LiveCache.Live.AudioType==0{//视频类直播
				if rankInfo.IsTop==0 && rankInfo.HopeIndex<=0&&rankInfo.Level<=50{
					videoList[dataInfo.LiveCache.Live.UserId]=1
				}
			}
		}
		for index := 0; index < ctx.GetDataLength(); index++ {
			dataInfo := ctx.GetDataByIndex(index).(*LiveInfo)
			rankInfo := dataInfo.GetRankInfo()
			if _,ok :=videoList[dataInfo.LiveCache.Live.UserId];ok{
				rankInfo.HopeIndex= abtest.GetInt("video_start_index",3)
				count+=1
			}
			if count>= abtest.GetInt("video_max_expo",8){
				break
			}
		}
	}

	return nil
	}
// 融合老策略的分数
type OldScoreStrategy struct{}

func (self *OldScoreStrategy) Do(ctx algo.IContext) error {
	var err error
	new_score := ctx.GetAbTest().GetFloat("new_score", 1.0)
	old_score := 1 - new_score
	for i := 0; i < ctx.GetDataLength(); i++ {
		dataInfo := ctx.GetDataByIndex(i)
		live := dataInfo.(*LiveInfo)
		rankInfo := dataInfo.GetRankInfo()
		score := self.oldScore(live)
		rankInfo.Score = live.RankInfo.Score*new_score + score*old_score
	}
	return err
}
func (self *OldScoreStrategy) scoreFx(score float32) float32 {
	return (score / 200) / (1 + score/200)
}
func (self *OldScoreStrategy) oldScore(live *LiveInfo) float32 {
	var score float32 = 0
	score += self.scoreFx(live.LiveCache.DayIncoming) * 0.2
	score += self.scoreFx(live.LiveCache.MonthIncoming) * 0.05
	score += self.scoreFx(live.LiveCache.Score) * 0.55
	score += self.scoreFx(float32(live.LiveCache.FansCount)) * 0.10
	score += self.scoreFx(float32(live.LiveCache.Live.ShareCount)) * 0.10
	return score
}

// 融合老策略的分数
type NewScoreStrategyV2 struct{}

func (self *NewScoreStrategyV2) Do(ctx algo.IContext) error {
	var err error
	algo_score := ctx.GetAbTest().GetFloat("algo_ratio", 0.3)
	business_score := 1 - algo_score
	for i := 0; i < ctx.GetDataLength(); i++ {
		dataInfo := ctx.GetDataByIndex(i)
		live := dataInfo.(*LiveInfo)
		rankInfo := dataInfo.GetRankInfo()
		score := self.oldScore(live)
		rankInfo.Score = live.RankInfo.Score*algo_score + score*business_score
	}
	return err
}
func (self *NewScoreStrategyV2) scoreFx(score float32) float32 {
	return utils.Expit(score / 600)
}
func (self *NewScoreStrategyV2) oldScore(live *LiveInfo) float32 {
	var score float32 = 0
	score += self.scoreFx(live.LiveCache.DayIncoming) * 0.5                //日收入
	score += self.scoreFx(live.LiveCache.MonthIncoming) * 0.25             //月收入
	score += self.scoreFx(live.LiveCache.Score) * 0.2                      //当前观看人数
	score += self.scoreFx(float32(live.LiveCache.FansCount)) * 0.025       //粉丝数
	score += self.scoreFx(float32(live.LiveCache.Live.ShareCount)) * 0.025 //分享数
	return score
}

type NewLiveStrategy struct{}

func (self *NewLiveStrategy) Do(ctx algo.IContext) error {
	var err error
	new_score := ctx.GetAbTest().GetFloat("algo_score", 0.1)
	old_score := 1 - new_score
	for i := 0; i < ctx.GetDataLength(); i++ {
		dataInfo := ctx.GetDataByIndex(i)
		live := dataInfo.(*LiveInfo)
		rankInfo := dataInfo.GetRankInfo()
		score := self.oldScore(live)
		rankInfo.Score = live.RankInfo.Score*new_score + score*old_score
	}
	return err
}
func (self *NewLiveStrategy) scoreFx(score float32) float32 {
	return (score / 200) / (1 + score/200)
}
func (self *NewLiveStrategy) oldScore(live *LiveInfo) float32 {
	var score float32 = 0
	score += self.scoreFx(live.LiveCache.Score) * 0.5
	score += self.scoreFx(float32(live.LiveCache.FansCount)) * 0.25
	score += self.scoreFx(float32(live.LiveCache.Live.ShareCount)) * 0.25
	return score
}

// 对于上个小时榜前3名进行随机制前
func HourRankRecommendFunc(ctx algo.IContext) error {
	abtest := ctx.GetAbTest()
	var startIndex = 5
	var intervar = 0
	params := ctx.GetRequest()
	topN := abtest.GetInt("per_hour_rank_top_n", 3) // 前n名随机， 分数相同的并列，有可能返回1,2,2,3
	indexs := []int{}
	for index := 0; index < ctx.GetDataLength(); index++ {
		dataInfo := ctx.GetDataByIndex(index).(*LiveInfo)
		userInfo := ctx.GetUserInfo().(*UserInfo)
		rankInfo := dataInfo.GetRankInfo()
		label :=0
		if dataInfo.LiveData != nil && dataInfo.LiveData.PreHourRank > 0 && dataInfo.LiveData.PreHourRank <= topN {
			indexs = append(indexs, index)
			label =1
			//continue //有上小时top3标签即不添加其他标签
		}
		if dataInfo.LiveCache.IsShowAdd == 1 {
			distance := rutils.EarthDistance(float64(params.Lng), float64(params.Lat), float64(dataInfo.LiveCache.Lng), float64(dataInfo.LiveCache.Lat))
			switch {
			case distance < 30000:
				if label==0{
					rankInfo.HopeIndex = startIndex + intervar*5
					intervar += 1
					label =1
				}
				dataInfo.LiveData.AddLabel(&labelItem{
					Style: AroundLabel,
					NewStyle:newStyle{
						Font:       "",
						Background: "https://static.rela.me/whitetag2",
						Color:      "313333",
					},
					Title: multiLanguage{
						Chs: "在你附近",
						Cht: "在你附近",
						En:  "Nearby",
					},
					weight: AroundWeight,
					level:  level1,
				})
			case distance >= 30000 && distance < 50000:
				if label==0{
					rankInfo.HopeIndex = startIndex + intervar*5
					intervar += 1
					label =1
				}
				dataInfo.LiveData.AddLabel(&labelItem{
					Style: CityLabel,
					NewStyle:newStyle{
						Font:       "",
						Background: "https://static.rela.me/whitetag2",
						Color:      "313333",
					},
					Title: multiLanguage{
						Chs: "同城",
						Cht: "同城",
						En:  "Local",
					},
					weight: CityWeight,
					level:  level1,
				})
			}
		}
		if userInfo.UserConcerns != nil {
			if userInfo.UserConcerns.Contains(dataInfo.UserId) {
				dataInfo.LiveData.AddLabel(&labelItem{
					Style: FollowLabel,
					NewStyle:newStyle{
						Font:       "",
						Background: "https://static.rela.me/whitetag2",
						Color:      "313333",
					},
					Title: multiLanguage{
						Chs: "你的关注",
						Cht: "你的关注",
						En:  "Following",
					},
					weight: FollowLabelWeight,
					level:  level1,
				})
			}
		}
		if userInfo.UserInterests != nil {
			if userInfo.UserInterests.Contains(dataInfo.UserId) {
				dataInfo.LiveData.AddLabel(&labelItem{
					Style: StrategyLabel,
					Title: multiLanguage{
						Chs: "猜你喜欢",
						Cht: "猜你喜歡",
						En:  "Recommended",
					},
					weight: StrategyLabelWeight,
					level:  level1,
				})
				if label==0{
					rankInfo.HopeIndex = startIndex + intervar*5
					intervar += 1
				}
			}
		}

	}
	if len(indexs) > 0 {
		i := rand.Intn(len(indexs))
		index := indexs[i]
		liveInfo := ctx.GetDataByIndex(index).(*LiveInfo)
		rankInfo := liveInfo.GetRankInfo()
		rankInfo.Level = rankInfo.Level + 99
		rankInfo.AddRecommendNeedReturn("PER_HOUR_TOP3", 2.0)
		liveInfo.LiveData.AddLabel(&labelItem{
			Style: HourRankLabel,
			NewStyle:newStyle{
				Font:       "",
				Background: "https://static.rela.me/yellotag.jpg",
				Color:      "ffffff",
			},
			Title: multiLanguage{
				Chs: "上小时TOP3",
				Cht: "上小時TOP3",
				En:  "TOP3",
			},
			weight: HourRankLabelWeight,
			level:  level1,
		})
	}
	return nil
}

func StrategyRecommendFunc(ctx algo.IContext) error {
	var startIndex = 5
	var intervar = 0
	params := ctx.GetRequest()
	for index := 0; index < ctx.GetDataLength(); index++ {
		dataInfo := ctx.GetDataByIndex(index).(*LiveInfo)
		userInfo := ctx.GetUserInfo().(*UserInfo)
		rankInfo := dataInfo.GetRankInfo()
		if dataInfo.LiveCache.IsShowAdd == 1 {
			distance := rutils.EarthDistance(float64(params.Lng), float64(params.Lat), float64(dataInfo.LiveCache.Lng), float64(dataInfo.LiveCache.Lat))
			switch {
			case distance < 30000:
				dataInfo.LiveData.AddLabel(&labelItem{
					Style: AroundLabel,
					Title: multiLanguage{
						Chs: "在你附近",
						Cht: "在你附近",
						En:  "Nearby",
					},
					weight: AroundWeight,
					level:  level1,
				})
			case distance >= 30000 && distance < 50000:
				dataInfo.LiveData.AddLabel(&labelItem{
					Style: CityLabel,
					Title: multiLanguage{
						Chs: "同城",
						Cht: "同城",
						En:  "Local",
					},
					weight: CityWeight,
					level:  level1,
				})
			}
		}
		if userInfo.UserConcerns != nil {
			if userInfo.UserConcerns.Contains(dataInfo.UserId) {
				dataInfo.LiveData.AddLabel(&labelItem{
					Style: FollowLabel,
					NewStyle:newStyle{
						Font:       "",
						Background: "https://static.rela.me/whitetag2",
						Color:      "313333",
					},
					Title: multiLanguage{
						Chs: "你的关注",
						Cht: "你的关注",
						En:  "Following",
					},
					weight: FollowLabelWeight,
					level:  level1,
				})
			}
		}
		if userInfo.UserInterests != nil {
			if userInfo.UserInterests.Contains(dataInfo.UserId) {
				rankInfo.HopeIndex = startIndex + intervar*7
				dataInfo.LiveData.AddLabel(&labelItem{
					Style: StrategyLabel,
					Title: multiLanguage{
						Chs: "猜你喜欢",
						Cht: "猜你喜歡",
						En:  "Recommended",
					},
					weight: StrategyLabelWeight,
					level:  level1,
				})
				intervar += 1
			}
		}
	}
	return nil
}

// 曝光未点击的直播降权 曝光两次未点击降权70%，曝光三次未点击降权50%，曝光四次未点击降权权30%，曝光五次以上未点击降权10%
func UserBehaviorExposureDownItemFunc(ctx algo.IContext, iDataInfo algo.IDataInfo, rankInfo *algo.RankInfo) error {
	dataInfo := iDataInfo.(*LiveInfo)
	if userBehavior := dataInfo.UserItemBehavior; userBehavior != nil {
		exposure := userBehavior.GetLiveExposure()
		if exposure.Count > 0 {
			if exposure.Count == 2 {
				rankInfo.AddRecommend("exposureDown", 0.7)
			} else if exposure.Count == 3 {
				rankInfo.AddRecommend("exposureDown", 0.5)
			} else if exposure.Count == 4 {
				rankInfo.AddRecommend("exposureDown", 0.3)
			} else if exposure.Count > 4 {
				rankInfo.AddRecommend("exposureDown", 0.1)
			}
		}
	}
	return nil
}

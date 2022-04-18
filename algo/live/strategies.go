package live

import (
	"math"
	"math/rand"
	"rela_recommend/algo"
	"rela_recommend/algo/utils"
	"rela_recommend/log"
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
	topN := abtest.GetInt("per_hour_rank_top_n", 3) // 前n名随机， 分数相同的并列，有可能返回1,2,2,3
	indexs := []int{}
	for index := 0; index < ctx.GetDataLength(); index++ {
		dataInfo := ctx.GetDataByIndex(index).(*LiveInfo)
		if dataInfo.LiveData != nil && dataInfo.LiveData.PreHourRank > 0 && dataInfo.LiveData.PreHourRank <= topN {
			indexs = append(indexs, index)
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
	var startIndex =1
	var intervar= 0
	for index := 0; index < ctx.GetDataLength(); index++ {
		dataInfo := ctx.GetDataByIndex(index).(*LiveInfo)
		userInfo := ctx.GetUserInfo().(*UserInfo)
		rankInfo := dataInfo.GetRankInfo()
		log.Warnf("userInfo,%s",userInfo)
		if userInfo.UserConcerns!=nil{
			log.Warnf("Userconcerns,%s,%s",userInfo.UserConcerns)
			log.Warnf("userid",dataInfo.UserId)
			if userInfo.UserConcerns.Contains(dataInfo.UserId){
				rankInfo.HopeIndex=startIndex+intervar*2
				dataInfo.LiveData.AddLabel(&labelItem{
					Style: FollowLabel,
					Title: multiLanguage{
						Chs: "你的关注",
						Cht: "你的关注",
						En:  "YOUR FOLLOW",
					},
					weight: FollowLabelWeight,
					level:  level1,
				})
			}
		}
		if userInfo.UserInterests!=nil{
			if userInfo.UserInterests.Contains(dataInfo.UserId){
				rankInfo.HopeIndex=startIndex+intervar*3
				dataInfo.LiveData.AddLabel(&labelItem{
					Style: StrategyLabel,
					Title: multiLanguage{
						Chs: "猜你喜欢",
						Cht: "猜你喜歡",
						En:  "GUESS YOU LIKE",
					},
					weight: StrategyLabelWeight,
					level:  level1,
				})
			}
		}
		intervar+=1
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

package live

import (
	"rela_recommend/algo/utils"
	rutils "rela_recommend/utils"
)

func GetLiveFeatures(ctx *LiveAlgoContext, live *LiveInfo) *utils.Features {
	fs := &utils.Features{}

	userCache := ctx.User.UserCache
	liveCache := live.LiveCache
	liveUserCache := live.UserCache
	// 用户连续特征 1 - 500
	fs.Add(1, float32(userCache.Age))
	fs.Add(2, float32(userCache.Height))
	fs.Add(3, float32(userCache.Weight))

	// 主播连续特征 501 - 1000
	fs.Add(501, float32(liveUserCache.Age))
	fs.Add(502, float32(liveUserCache.Height))
	fs.Add(503, float32(liveUserCache.Weight))
	fs.Add(504, float32(ctx.CreateTime.Sub(liveUserCache.CreateTime.Time).Seconds() / 60 / 60 / 24))			// 直播注册时长 day
	fs.Add(505, float32(liveCache.FansCount))		// 粉丝数量
	fs.Add(506, liveCache.DayIncoming)				// 天收入
	fs.Add(507, liveCache.MonthIncoming)			// 月收入
	fs.Add(508, float32(liveCache.Live.ShareCount))			// 分享次数
	fs.Add(509, float32(liveCache.Live.SendMsgCount))			// 消息数量
	fs.Add(510, float32(liveCache.Live.GemProfit))			// 房间收益
	fs.Add(511, float32(ctx.CreateTime.Sub(liveCache.Live.CreateTime.Time).Seconds() / 60 / 60 / 24))			// 房间开播时长 min

	// 离散特征  1000 - &
	fs.AddCategory(10000, 24, 0, ctx.CreateTime.Hour(), 0)			// 时间
	fs.AddCategory(10024, 7, 0, int(ctx.CreateTime.Weekday()), 0)	// 周几

	// 用户
	fs.AddCategory(10100, 12, 0, rutils.GetInt(userCache.Horoscope), 0)	// 星座
	fs.AddCategory(10120, 10, -1, userCache.Affection, -1)				// 单身情况
	fs.AddCategory(10130, 10, 0, rutils.GetInt(userCache.RoleName), 0)				// 自我认同
	fs.AddCategories(10140, 10, 0, rutils.GetInts(userCache.WantRole), 0)	// 想要寻找
	
	// 主播
	fs.AddCategory(10150, 12, 0, rutils.GetInt(liveUserCache.Horoscope), 0)	// 星座
	fs.AddCategory(10170, 10, -1, liveUserCache.Affection, -1)				// 单身情况
	fs.AddCategory(10180, 10, 0, rutils.GetInt(liveUserCache.RoleName), 0)				// 自我认同
	fs.AddCategories(10190, 10, 0, rutils.GetInts(liveUserCache.WantRole), 0)	// 想要寻找

	fs.AddCategory(10200, 5, 0, ctx.Platform, 0)	// 用户操作系统
	fs.AddCategory(10205, 5, 0, rutils.GetPlatform(liveCache.Live.Ua), 0)	// 主播操作系统

	fs.AddCategory(10210, 10, 0, liveCache.Live.AudioType, 0)	// 房间类型
	fs.AddCategory(10220, 10, 0, liveCache.Live.IsMulti, 0)	// 房间是否多人

	fs.AddCategory(10230, 2, 0, rutils.GetInt(ctx.User.UserConcerns.Contains(liveCache.Live.UserId)), 0)	// 用户是否是主播粉丝

	return fs
}



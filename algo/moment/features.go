package moment

import (
	"rela_recommend/factory"
	"rela_recommend/algo/utils"
	rutils "rela_recommend/utils"
)

func GetMomentFeatures(model IMomentAlgo, ctx *AlgoContext, data *DataInfo) *utils.Features {
	fs := &utils.Features{}
	shareToList := map[string]int{"all": 0, "friends": 1, "self": 2}

	// 发布内容
	mem, meme := data.MomentCache, data.MomentExtendCache
	memprofile := data.MomentProfile
	fs.AddCategory(1, 12, 0, mem.InsertTime.Hour(), 0)
	fs.AddCategory(30, 7, 0, int(mem.InsertTime.Weekday()), 0)
	if shareTo, ok := shareToList[mem.ShareTo]; ok {
		fs.AddCategory(40, 3, 0, shareTo, 0)
	}
	// 		是否有 内容，图片，视频
	wordsCount := len(mem.MomentsText)
	fs.AddCategory(44, 2, 0, rutils.GetInt(wordsCount > 0), 0)
	fs.AddCategory(46, 2, 0, rutils.GetInt(len(mem.ImageUrl) > 0), 0)
	fs.AddCategory(48, 2, 0, rutils.GetInt(len(mem.VoiceUrl) > 0), 0)

	if memprofile != nil {
		fs.Add(50, float32(memprofile.TextCnt))
		fs.Add(51, float32(memprofile.LikeCnt))
	}
	fs.Add(52, float32(len(mem.MomentsText)))

	// 发布者
	memu := data.UserCache
	var role, wantRoles = 0, make([]int, 0)
	if(memu != nil) {
		fs.Add(2000, float32(ctx.CreateTime.Sub(memu.CreateTime.Time).Seconds() / 60 / 60 / 24))
		fs.Add(2001, float32(memu.Age))
		fs.Add(2002, float32(memu.Height))
		fs.Add(2003, float32(memu.Weight))

		fs.AddCategory(2010, 13, -1, rutils.GetInt(memu.Horoscope), -1)	// 星座
		fs.AddCategory(2030, 10, -1, memu.Affection, -1)				// 单身情况
		role, wantRoles = rutils.GetInt(memu.RoleName), rutils.GetInts(memu.WantRole)
		fs.AddCategory(2040, 10, -1, role, -1)				// 自我认同
		fs.AddCategories(2050, 10, -1, wantRoles, -1)	// 想要寻找
	}

	// 观看者
	if ctx.User != nil && ctx.User.UserCache != nil {
		curr := ctx.User.UserCache
		fs.Add(4000, float32(ctx.CreateTime.Sub(curr.CreateTime.Time).Seconds() / 60 / 60 / 24))
		fs.Add(4001, float32(curr.Age))
		fs.Add(4002, float32(curr.Height))
		fs.Add(4003, float32(curr.Weight))
		fs.Add(4004, float32(rutils.EarthDistance(float64(ctx.Request.Lng), float64(ctx.Request.Lat), meme.Lng, meme.Lat)))
		fs.Add(4005, float32(ctx.CreateTime.Sub(mem.InsertTime).Minutes()))

		fs.AddCategory(4010, 13, -1, rutils.GetInt(curr.Horoscope), -1)	// 星座
		fs.AddCategory(4030, 10, -1, curr.Affection, -1)				// 单身情况
		uRole, uWantRoles := rutils.GetInt(curr.RoleName), rutils.GetInts(curr.WantRole)
		fs.AddCategory(4040, 10, -1, uRole, -1)				// 自我认同
		fs.AddCategories(4050, 10, -1, uWantRoles, -1)	// 想要寻找

		// 交叉
		fs.AddCategory(6000, 2, 0, rutils.GetInt(rutils.IsInInts(role, uWantRoles)), 0)
		fs.AddCategory(6002, 2, 0, rutils.GetInt(rutils.IsInInts(uRole, wantRoles)), 0)
	}
	
	// 分词结果
	if wordsCount > 0 {
		var words = make([]string, 0)
		if memprofile != nil {
			words = memprofile.MomentsTextWords
		} else {
			words = factory.Segmenter.Cut(mem.MomentsText)
		}
		words = model.CheckWords(words)
		fs.AddHashStrings(100000, 100000, words)
	}
	return fs
}

package moment

import (
	"rela_recommend/algo"
	"rela_recommend/algo/utils"
	"rela_recommend/factory"
	rutils "rela_recommend/utils"
	"strings"
	"time"
)

func GetMomLabel(label string,labelMap map[string]int) []int{
	var ids = make([]int, 0)
	if label !=""{
		for _, uid := range strings.Split(label, ",") {
			if val,ok :=labelMap[uid];ok{
				ids = append(ids, rutils.GetInt(val))
			}
		}
		return ids
	}
	return ids
}

func GetMomentFeatures(ctx algo.IContext, model algo.IAlgo, idata algo.IDataInfo) *utils.Features {
	fs := &utils.Features{}
	shareToList := map[string]int{"all": 0, "friends": 1, "self": 2}
	momsTypeList :=map[string]int{"text":0,"text_image":1,"image":2,"live":3,"theme":4,"themereply":5,"video":6,"voice_live":7,"ad":8,"recommend":9}
	momsPicLabelList :=map[string]int{"卡通人脸":0,"明星人脸":1,"自拍":2,"人脸":3,"人":4,"logo":5,"美食":6,"风景":7,"表情包":8,"动漫":9,"游戏":10,"宠物":11,"运动":12,"时尚":13,"模糊":14}
	currTime := ctx.GetCreateTime().Unix()
	data := idata.(*DataInfo)
	// 发布内容
	mem, meme := data.MomentCache, data.MomentExtendCache
	memprofile := data.MomentProfile
	if mem != nil {
		fs.AddCategory(1, 24, 0, mem.InsertTime.Hour(), 0)
		fs.AddCategory(30, 7, 0, int(mem.InsertTime.Weekday()), 0)
		if shareTo, ok := shareToList[mem.ShareTo]; ok {
			fs.AddCategory(40, 3, 0, shareTo, 0)
		}
		// 		是否有 内容，图片，视频
		wordsCount := len(mem.MomentsText)
		fs.AddCategory(44, 2, 0, rutils.GetInt(wordsCount > 0), 0)
		fs.AddCategory(46, 2, 0, rutils.GetInt(len(mem.ImageUrl) > 0), 0)
		fs.AddCategory(48, 2, 0, rutils.GetInt(len(mem.VoiceUrl) > 0), 0)
		if (len(mem.VoiceUrl) > 0) {
			fs.Add(54, float32(len(strings.Split(mem.ImageUrl, ","))))
		}
		if memprofile != nil {
			fs.Add(50, float32(memprofile.TextCnt))
			fs.Add(51, float32(memprofile.LikeCnt))
		}
		fs.Add(52, float32(len(mem.MomentsText)))
		fs.Add(53, float32(ctx.GetCreateTime().Sub(mem.InsertTime).Minutes()))
		fs.Add(55, float32(time.Now().Hour()))
		//日志类型
		if momsType, ok := momsTypeList[mem.MomentsType]; ok {
			fs.AddCategory(70, 10, 0, momsType, 0)
		}


		//日志离线画像
		momOfflineProfile := data.MomentOfflineProfile
		if (momOfflineProfile != nil) {
			fs.AddArray(100, 128, momOfflineProfile.MomentEmbedding)
		}

		//日志内容画像  1000-。。。
		momContentProfile :=data.MomentContentProfile
		if momContentProfile!=nil{
			fs.AddCategories(1000,15,0,GetMomLabel(momContentProfile.Tags,momsPicLabelList),0)
		}
		// 发布者
		memu := data.UserCache
		var role, wantRoles = 0, make([]int, 0)
		if (memu != nil) {
			fs.Add(2000, float32(ctx.GetCreateTime().Sub(memu.CreateTime.Time).Seconds()/60/60/24))
			fs.Add(2001, float32(memu.Age))
			fs.Add(2002, float32(memu.Height))
			fs.Add(2003, float32(memu.Weight))

			fs.AddCategory(2010, 13, -1, rutils.GetInt(memu.Horoscope), -1) // 星座
			fs.AddCategory(2030, 10, -1, memu.Affection, -1)                // 单身情况
			role, wantRoles = rutils.GetInt(memu.RoleName), rutils.GetInts(memu.WantRole)
			fs.AddCategory(2040, 10, -1, role, -1)        // 自我认同
			fs.AddCategories(2050, 10, -1, wantRoles, -1) // 想要寻找
			fs.Add(2100,float32(memu.MomentsCount))//发布者历史发布日志数
			fs.Add(2101,float32(memu.Grade))//优质用户评分 非优质用户0
		}
		memuEmbedding := data.MomentUserProfile
		if (memuEmbedding != nil) {
			fs.AddArray(3000, 128, memuEmbedding.UserEmbedding)
		}

		// 观看者
		if ctx.GetUserInfo() != nil {
			user := ctx.GetUserInfo().(*UserInfo)
			if user.UserCache != nil {
				curr := user.UserCache
				fs.Add(4000, float32(ctx.GetCreateTime().Sub(curr.CreateTime.Time).Seconds()/60/60/24))
				fs.Add(4001, float32(curr.Age))
				fs.Add(4002, float32(curr.Height))
				fs.Add(4003, float32(curr.Weight))
				if meme != nil {
					fs.Add(4004, float32(rutils.EarthDistance(float64(ctx.GetRequest().Lng), float64(ctx.GetRequest().Lat), meme.Lng, meme.Lat)))
				}
				fs.AddCategory(4010, 13, -1, rutils.GetInt(curr.Horoscope), -1) // 星座
				fs.AddCategory(4030, 10, -1, curr.Affection, -1)                // 单身情况
				uRole, uWantRoles := rutils.GetInt(curr.RoleName), rutils.GetInts(curr.WantRole)
				fs.AddCategory(4040, 10, -1, uRole, -1)        // 自我认同
				fs.AddCategories(4050, 10, -1, uWantRoles, -1) // 想要寻找
				// 交叉
				fs.AddCategory(6000, 2, 0, rutils.GetInt(rutils.IsInInts(role, uWantRoles)), 0)
				fs.AddCategory(6002, 2, 0, rutils.GetInt(rutils.IsInInts(uRole, wantRoles)), 0)
			}
			if user.MomentUserProfile != nil {
				fs.AddArray(5100, 128, user.MomentUserProfile.UserEmbedding)
				fs.AddArray(7000, 128, user.MomentUserProfile.FollowEmbedding)
			}
			if memuEmbedding != nil && user.MomentUserProfile != nil {
				fs.Add(6100, utils.ArrayMultSum(memuEmbedding.UserEmbedding, user.MomentUserProfile.UserEmbedding))
			}

		} // 该内容实时行为特征
		if data.ItemBehavior != nil {
			// 互动
			listInteract := data.ItemBehavior.GetMomentListInteract()
			//点击
			listClick :=data.ItemBehavior.GetMomentListClick()
			fs.Add(8999,float32(listClick.Count))
			fs.Add(9000, float32(listInteract.Count))
			if listInteract.LastTime > 0 {
				fs.Add(9001, float32(float64(currTime)-listInteract.LastTime))
			}
			// 曝光
			listExposure := data.ItemBehavior.GetMomentListExposure()
			fs.Add(9002, float32(listExposure.Count))
			if listExposure.LastTime > 0 {
				fs.Add(9003, float32(float64(currTime)-listExposure.LastTime))
				fs.Add(9004, float32(listInteract.Count/listExposure.Count)) // 互动率
			}
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
	}
	return fs
}

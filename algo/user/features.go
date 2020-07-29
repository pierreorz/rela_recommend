package user

import (
	"rela_recommend/algo"
	"rela_recommend/algo/utils"
	"rela_recommend/models/redis"
	rutils "rela_recommend/utils"
)

func GetFeaturesV0(ctx algo.IContext, model algo.IAlgo, idata algo.IDataInfo) *utils.Features {
	fs := &utils.Features{}
	data := idata.(*DataInfo)
	currTime := ctx.GetCreateTime().Unix()

	// 用户
	var role, wantRoles = 0, make([]int, 0)
	var userCache *redis.UserProfile
	var userProfile *redis.NearbyProfile
	if ctx.GetUserInfo() != nil {
		user := ctx.GetUserInfo().(*UserInfo)
		if user.UserCache != nil {
			userCache = user.UserCache
			fs.Add(1, float32(userCache.Age))
			fs.Add(2, float32(userCache.Height))
			fs.Add(3, float32(userCache.Weight))
			fs.Add(4, float32(currTime-userCache.CreateTime.Time.Unix()))
			fs.AddCategory(10, 15, -1, rutils.GetInt(userCache.Horoscope), -1)
			fs.AddCategory(30, 10, -1, userCache.Affection, -1)
			// role, wantRoles = rutils.GetInt(userCache.RoleName), rutils.GetInts(userCache.WantRole)
			role, wantRoles = userCache.GetRoleNameInt(), userCache.GetWantRoleInts()
			fs.AddCategory(50, 15, -1, role, -1)
			fs.AddCategories(70, 15, 0, wantRoles, 0)
		}

		// 用户画像
		userProfile = user.UserProfile
		if userProfile != nil {
			if userProfile.AgeMap != nil {
				fs.Add(2000, userProfile.AgeMap["age_18_20"])
				fs.Add(2001, userProfile.AgeMap["age_21_22"])
				fs.Add(2002, userProfile.AgeMap["age_23_24"])
				fs.Add(2003, userProfile.AgeMap["age_25_26"])
				fs.Add(2004, userProfile.AgeMap["age_27_29"])
				fs.Add(2005, userProfile.AgeMap["age_30_40"])
				fs.Add(2006, userProfile.AgeMap["age_above_40"])
				fs.Add(2007, userProfile.AgeMap["age_unknown"])
			}
			if userProfile.RoleNameMap != nil {
				fs.Add(2010, userProfile.RoleNameMap["role_name_t"])
				fs.Add(2011, userProfile.RoleNameMap["role_name_p"])
				fs.Add(2012, userProfile.RoleNameMap["role_name_h"])
				fs.Add(2013, userProfile.RoleNameMap["role_name_bi"])
				fs.Add(2014, userProfile.RoleNameMap["role_name_other"])
				fs.Add(2015, userProfile.RoleNameMap["role_name_str"])
				fs.Add(2016, userProfile.RoleNameMap["role_name_fu"])
				fs.Add(2017, userProfile.RoleNameMap["role_name_unknown"])
			}
			if userProfile.HoroscopeMap != nil {
				fs.Add(2020, userProfile.HoroscopeMap["horoscope_cap"])
				fs.Add(2021, userProfile.HoroscopeMap["horoscope_aqua"])
				fs.Add(2022, userProfile.HoroscopeMap["horoscope_pis"])
				fs.Add(2023, userProfile.HoroscopeMap["horoscope_ar"])
				fs.Add(2024, userProfile.HoroscopeMap["horoscope_tau"])
				fs.Add(2025, userProfile.HoroscopeMap["horoscope_gemini"])
				fs.Add(2026, userProfile.HoroscopeMap["horoscope_cancer"])
				fs.Add(2027, userProfile.HoroscopeMap["horoscope_leo"])
				fs.Add(2028, userProfile.HoroscopeMap["horoscope_virgo"])
				fs.Add(2029, userProfile.HoroscopeMap["horoscope_libra"])
				fs.Add(2030, userProfile.HoroscopeMap["horoscope_scor"])
				fs.Add(2031, userProfile.HoroscopeMap["horoscope_sagi"])
				fs.Add(2032, userProfile.HoroscopeMap["horoscope_unknown"])
			}
			if userProfile.HeightMap != nil {
				fs.Add(2040, userProfile.HeightMap["height_under_155"])
				fs.Add(2041, userProfile.HeightMap["height_156_160"])
				fs.Add(2042, userProfile.HeightMap["height_161_163"])
				fs.Add(2043, userProfile.HeightMap["height_164_166"])
				fs.Add(2044, userProfile.HeightMap["height_167_170"])
				fs.Add(2045, userProfile.HeightMap["height_171_180"])
				fs.Add(2046, userProfile.HeightMap["height_above_180"])
				fs.Add(2047, userProfile.HeightMap["height_unknown"])
			}
			if userProfile.WeightMap != nil {
				fs.Add(2050, userProfile.WeightMap["weight_under_41"])
				fs.Add(2051, userProfile.WeightMap["weight_42_45"])
				fs.Add(2052, userProfile.WeightMap["weight_46_49"])
				fs.Add(2053, userProfile.WeightMap["weight_50_52"])
				fs.Add(2054, userProfile.WeightMap["weight_53_57"])
				fs.Add(2055, userProfile.WeightMap["weight_above_58"])
				fs.Add(2056, userProfile.WeightMap["weight_unknown"])
			}
			if userProfile.DistanceMap != nil {
				fs.Add(2060, userProfile.DistanceMap["dis_unknown"])
				fs.Add(2061, userProfile.DistanceMap["dis_0_03"])
				fs.Add(2062, userProfile.DistanceMap["dis_03_1"])
				fs.Add(2063, userProfile.DistanceMap["dis_1_5"])
				fs.Add(2064, userProfile.DistanceMap["dis_5_20"])
				fs.Add(2065, userProfile.DistanceMap["dis_20_40"])
				fs.Add(2066, userProfile.DistanceMap["dis_40_70"])
				fs.Add(2067, userProfile.DistanceMap["dis_70_100"])
				fs.Add(2068, userProfile.DistanceMap["dis_100_300"])
				fs.Add(2069, userProfile.DistanceMap["dis_300_500"])
				fs.Add(2070, userProfile.DistanceMap["dis_above_500"])
			}
			if userProfile.AffectionMap != nil {
				fs.Add(2080, userProfile.AffectionMap["affection_unknown"])
				fs.Add(2081, userProfile.AffectionMap["affection_single"])
				fs.Add(2082, userProfile.AffectionMap["affection_dating"])
				fs.Add(2083, userProfile.AffectionMap["affection_stable"])
				fs.Add(2084, userProfile.AffectionMap["affection_married"])
				fs.Add(2085, userProfile.AffectionMap["affection_open_re"])
				fs.Add(2086, userProfile.AffectionMap["affection_relationship"])
				fs.Add(2087, userProfile.AffectionMap["affection_waiting"])
				fs.Add(2088, userProfile.AffectionMap["affection_secret"])
			}
			if userProfile.MobileSysMap != nil {
				fs.Add(2090, userProfile.MobileSysMap["mobile_sys_ios"])
				fs.Add(2091, userProfile.MobileSysMap["mobile_sys_android"])
			}
			if userProfile.FreqWeekMap != nil {
				fs.Add(2100, userProfile.FreqWeekMap["monday"])
				fs.Add(2101, userProfile.FreqWeekMap["tuesday"])
				fs.Add(2102, userProfile.FreqWeekMap["wednesday"])
				fs.Add(2103, userProfile.FreqWeekMap["thursday"])
				fs.Add(2104, userProfile.FreqWeekMap["friday"])
				fs.Add(2105, userProfile.FreqWeekMap["saturday"])
				fs.Add(2106, userProfile.FreqWeekMap["sunday"])
			}
			if userProfile.FreqTimeMap != nil {
				fs.Add(2110, userProfile.FreqTimeMap["time_0_2"])
				fs.Add(2111, userProfile.FreqTimeMap["time_2_4"])
				fs.Add(2112, userProfile.FreqTimeMap["time_4_6"])
				fs.Add(2113, userProfile.FreqTimeMap["time_6_8"])
				fs.Add(2114, userProfile.FreqTimeMap["time_8_10"])
				fs.Add(2115, userProfile.FreqTimeMap["time_10_12"])
				fs.Add(2116, userProfile.FreqTimeMap["time_12_14"])
				fs.Add(2117, userProfile.FreqTimeMap["time_14_16"])
				fs.Add(2118, userProfile.FreqTimeMap["time_16_18"])
				fs.Add(2119, userProfile.FreqTimeMap["time_18_20"])
				fs.Add(2110, userProfile.FreqTimeMap["time_20_22"])
				fs.Add(2121, userProfile.FreqTimeMap["time_22_24"])
			}
			if userProfile.TotalMap != nil {
				fs.Add(2130, userProfile.TotalMap["total_cnt"])
				fs.Add(2131, userProfile.TotalMap["total_day"])
				fs.Add(2132, userProfile.TotalMap["total_received_cnt"])
			}
			if userProfile.NearSeeMap != nil {
				fs.Add(2140, userProfile.NearSeeMap["near_see_cnt"])
				fs.Add(2141, userProfile.NearSeeMap["near_see_click_rate"])
				fs.Add(2142, userProfile.NearSeeMap["near_see_cnt_7d"])
				fs.Add(2143, userProfile.NearSeeMap["near_see_click_rate_7d"])
			}
			if userProfile.NearShowMap != nil {
				fs.Add(2150, userProfile.NearShowMap["near_show_cnt"])
				fs.Add(2151, userProfile.NearShowMap["near_show_click_rate"])
				fs.Add(2152, userProfile.NearShowMap["near_show_cnt_7d"])
				fs.Add(2153, userProfile.NearShowMap["near_show_click_rate_7d"])
			}
			if userProfile.ActiveTimeMap != nil {
				fs.Add(2160, userProfile.ActiveTimeMap["active_time_unknown"])
				fs.Add(2161, userProfile.ActiveTimeMap["active_time_1_60"])
				fs.Add(2162, userProfile.ActiveTimeMap["active_time_60_300"])
				fs.Add(2163, userProfile.ActiveTimeMap["active_time_300_1800"])
				fs.Add(2164, userProfile.ActiveTimeMap["active_time_1800_14400"])
				fs.Add(2165, userProfile.ActiveTimeMap["active_time_14400_86400"])
				fs.Add(2166, userProfile.ActiveTimeMap["active_time_1d_7d"])
				fs.Add(2167, userProfile.ActiveTimeMap["active_time_above_7d"])
			}
		}
	}

	curr := data.UserCache
	currProfile := data.UserProfile

	var cRole, cWantRoles = 0, make([]int, 0)
	if curr != nil {
		fs.Add(5000, float32(currTime-curr.LastUpdateTime))
		fs.Add(5001, float32(curr.Age))
		fs.Add(5002, float32(curr.Height))
		fs.Add(5003, float32(curr.Weight))
		fs.Add(5004, float32(currTime-curr.CreateTime.Time.Unix()))
		fs.AddCategory(5010, 15, -1, rutils.GetInt(curr.Horoscope), -1)
		fs.AddCategory(5030, 10, -1, curr.Affection, -1)
		// cRole, cWantRoles = rutils.GetInt(curr.RoleName), rutils.GetInts(curr.WantRole)
		cRole, cWantRoles = curr.GetRoleNameInt(), curr.GetWantRoleInts()
		fs.AddCategory(5050, 15, -1, cRole, -1) // 自我认同
		fs.AddCategories(5070, 15, -1, cWantRoles, -1)
		fs.AddCategory(5100, 2, 0, rutils.GetInt(data.LiveInfo != nil), 0) // 是否正在直播
	}

	if currProfile != nil {
		if currProfile.AgeMap != nil {
			fs.Add(7000, currProfile.AgeMap["age_18_20"])
			fs.Add(7001, currProfile.AgeMap["age_21_22"])
			fs.Add(7002, currProfile.AgeMap["age_23_24"])
			fs.Add(7003, currProfile.AgeMap["age_25_26"])
			fs.Add(7004, currProfile.AgeMap["age_27_29"])
			fs.Add(7005, currProfile.AgeMap["age_30_40"])
			fs.Add(7006, currProfile.AgeMap["age_above_40"])
			fs.Add(7007, currProfile.AgeMap["age_unknown"])
		}
		if currProfile.RoleNameMap != nil {
			fs.Add(7010, currProfile.RoleNameMap["role_name_t"])
			fs.Add(7011, currProfile.RoleNameMap["role_name_p"])
			fs.Add(7012, currProfile.RoleNameMap["role_name_h"])
			fs.Add(7013, currProfile.RoleNameMap["role_name_bi"])
			fs.Add(7014, currProfile.RoleNameMap["role_name_other"])
			fs.Add(7015, currProfile.RoleNameMap["role_name_str"])
			fs.Add(7016, currProfile.RoleNameMap["role_name_fu"])
			fs.Add(7017, currProfile.RoleNameMap["role_name_unknown"])
		}
		if currProfile.HoroscopeMap != nil {
			fs.Add(7020, currProfile.HoroscopeMap["horoscope_cap"])
			fs.Add(7021, currProfile.HoroscopeMap["horoscope_aqua"])
			fs.Add(7022, currProfile.HoroscopeMap["horoscope_pis"])
			fs.Add(7023, currProfile.HoroscopeMap["horoscope_ar"])
			fs.Add(7024, currProfile.HoroscopeMap["horoscope_tau"])
			fs.Add(7025, currProfile.HoroscopeMap["horoscope_gemini"])
			fs.Add(7026, currProfile.HoroscopeMap["horoscope_cancer"])
			fs.Add(7027, currProfile.HoroscopeMap["horoscope_leo"])
			fs.Add(7028, currProfile.HoroscopeMap["horoscope_virgo"])
			fs.Add(7029, currProfile.HoroscopeMap["horoscope_libra"])
			fs.Add(7030, currProfile.HoroscopeMap["horoscope_scor"])
			fs.Add(7031, currProfile.HoroscopeMap["horoscope_sagi"])
			fs.Add(7032, currProfile.HoroscopeMap["horoscope_unknown"])
		}
		if currProfile.HeightMap != nil {
			fs.Add(7040, currProfile.HeightMap["height_under_155"])
			fs.Add(7041, currProfile.HeightMap["height_156_160"])
			fs.Add(7042, currProfile.HeightMap["height_161_163"])
			fs.Add(7043, currProfile.HeightMap["height_164_166"])
			fs.Add(7044, currProfile.HeightMap["height_167_170"])
			fs.Add(7045, currProfile.HeightMap["height_171_180"])
			fs.Add(7046, currProfile.HeightMap["height_above_180"])
			fs.Add(7047, currProfile.HeightMap["height_unknown"])
		}
		if currProfile.WeightMap != nil {
			fs.Add(7050, currProfile.WeightMap["weight_under_41"])
			fs.Add(7051, currProfile.WeightMap["weight_42_45"])
			fs.Add(7052, currProfile.WeightMap["weight_46_49"])
			fs.Add(7053, currProfile.WeightMap["weight_50_52"])
			fs.Add(7054, currProfile.WeightMap["weight_53_57"])
			fs.Add(7055, currProfile.WeightMap["weight_above_58"])
			fs.Add(7056, currProfile.WeightMap["weight_unknown"])
		}
		if currProfile.DistanceMap != nil {
			fs.Add(7060, currProfile.DistanceMap["dis_unknown"])
			fs.Add(7061, currProfile.DistanceMap["dis_0_03"])
			fs.Add(7062, currProfile.DistanceMap["dis_03_1"])
			fs.Add(7063, currProfile.DistanceMap["dis_1_5"])
			fs.Add(7064, currProfile.DistanceMap["dis_5_20"])
			fs.Add(7065, currProfile.DistanceMap["dis_20_40"])
			fs.Add(7066, currProfile.DistanceMap["dis_40_70"])
			fs.Add(7067, currProfile.DistanceMap["dis_70_100"])
			fs.Add(7068, currProfile.DistanceMap["dis_100_300"])
			fs.Add(7069, currProfile.DistanceMap["dis_300_500"])
			fs.Add(7070, currProfile.DistanceMap["dis_above_500"])
		}
		if currProfile.AffectionMap != nil {
			fs.Add(7080, currProfile.AffectionMap["affection_unknown"])
			fs.Add(7081, currProfile.AffectionMap["affection_single"])
			fs.Add(7082, currProfile.AffectionMap["affection_dating"])
			fs.Add(7083, currProfile.AffectionMap["affection_stable"])
			fs.Add(7084, currProfile.AffectionMap["affection_married"])
			fs.Add(7085, currProfile.AffectionMap["affection_open_re"])
			fs.Add(7086, currProfile.AffectionMap["affection_relationship"])
			fs.Add(7087, currProfile.AffectionMap["affection_waiting"])
			fs.Add(7088, currProfile.AffectionMap["affection_secret"])
		}
		if currProfile.MobileSysMap != nil {
			fs.Add(7090, currProfile.MobileSysMap["mobile_sys_ios"])
			fs.Add(7091, currProfile.MobileSysMap["mobile_sys_android"])
		}
		if currProfile.FreqWeekMap != nil {
			fs.Add(7100, currProfile.FreqWeekMap["monday"])
			fs.Add(7101, currProfile.FreqWeekMap["tuesday"])
			fs.Add(7102, currProfile.FreqWeekMap["wednesday"])
			fs.Add(7103, currProfile.FreqWeekMap["thursday"])
			fs.Add(7104, currProfile.FreqWeekMap["friday"])
			fs.Add(7105, currProfile.FreqWeekMap["saturday"])
			fs.Add(7106, currProfile.FreqWeekMap["sunday"])
		}
		if currProfile.FreqTimeMap != nil {
			fs.Add(7110, currProfile.FreqTimeMap["time_0_2"])
			fs.Add(7111, currProfile.FreqTimeMap["time_2_4"])
			fs.Add(7112, currProfile.FreqTimeMap["time_4_6"])
			fs.Add(7113, currProfile.FreqTimeMap["time_6_8"])
			fs.Add(7114, currProfile.FreqTimeMap["time_8_10"])
			fs.Add(7115, currProfile.FreqTimeMap["time_10_12"])
			fs.Add(7116, currProfile.FreqTimeMap["time_12_14"])
			fs.Add(7117, currProfile.FreqTimeMap["time_14_16"])
			fs.Add(7118, currProfile.FreqTimeMap["time_16_18"])
			fs.Add(7119, currProfile.FreqTimeMap["time_18_20"])
			fs.Add(7110, currProfile.FreqTimeMap["time_20_22"])
			fs.Add(7121, currProfile.FreqTimeMap["time_22_24"])
		}
		if currProfile.TotalMap != nil {
			fs.Add(7130, currProfile.TotalMap["total_cnt"])
			fs.Add(7131, currProfile.TotalMap["total_day"])
			fs.Add(7132, currProfile.TotalMap["total_received_cnt"])
		}
		if currProfile.NearSeeMap != nil {
			fs.Add(7140, currProfile.NearSeeMap["near_see_cnt"])
			fs.Add(7141, currProfile.NearSeeMap["near_see_click_rate"])
			fs.Add(7142, currProfile.NearSeeMap["near_see_cnt_7d"])
			fs.Add(7143, currProfile.NearSeeMap["near_see_click_rate_7d"])
		}
		if currProfile.NearShowMap != nil {
			fs.Add(7150, currProfile.NearShowMap["near_show_cnt"])
			fs.Add(7151, currProfile.NearShowMap["near_show_click_rate"])
			fs.Add(7152, currProfile.NearShowMap["near_show_cnt_7d"])
			fs.Add(7153, currProfile.NearShowMap["near_show_click_rate_7d"])
		}
		if currProfile.ActiveTimeMap != nil {
			fs.Add(7160, currProfile.ActiveTimeMap["active_time_unknown"])
			fs.Add(7161, currProfile.ActiveTimeMap["active_time_1_60"])
			fs.Add(7162, currProfile.ActiveTimeMap["active_time_60_300"])
			fs.Add(7163, currProfile.ActiveTimeMap["active_time_300_1800"])
			fs.Add(7164, currProfile.ActiveTimeMap["active_time_1800_14400"])
			fs.Add(7165, currProfile.ActiveTimeMap["active_time_14400_86400"])
			fs.Add(7166, currProfile.ActiveTimeMap["active_time_1d_7d"])
			fs.Add(7167, currProfile.ActiveTimeMap["active_time_above_7d"])
		}
	}
	// 该内容实时行为特征
	if data.ItemBehavior != nil {
		// 点击互动
		listInteract := data.ItemBehavior.GetNearbyListInteract()
		fs.Add(9000, float32(listInteract.Count))
		if listInteract.LastTime > 0 {
			fs.Add(9001, float32(float64(currTime)-listInteract.LastTime))
		}
		// 曝光
		listExposure := data.ItemBehavior.GetNearbyListExposure()
		fs.Add(9002, float32(listExposure.Count))
		if listExposure.LastTime > 0 {
			fs.Add(9003, float32(float64(currTime)-listExposure.LastTime))
			fs.Add(9004, float32(listInteract.Count/listExposure.Count)) // 互动率
		}
	}
	// 该用户对内容实时行为特征
	if data.UserBehavior != nil {
		// 点击互动
		listInteract := data.UserBehavior.GetNearbyListInteract()
		fs.Add(9010, float32(listInteract.Count))
		if listInteract.LastTime > 0 {
			fs.Add(9011, float32(float64(currTime)-listInteract.LastTime))
		}
		// 曝光
		listExposure := data.UserBehavior.GetNearbyListExposure()
		fs.Add(9012, float32(listExposure.Count))
		if listExposure.LastTime > 0 {
			fs.Add(9013, float32(float64(currTime)-listExposure.LastTime))
			fs.Add(9014, float32(listInteract.Count/listExposure.Count)) // 互动率
		}

	}

	// ****************************************************    交叉特征
	fs.AddCategory(10000, 2, 0, rutils.GetInt(role > 0 && rutils.IsInInts(role, cWantRoles)), 0)
	fs.AddCategory(10002, 2, 0, rutils.GetInt(cRole > 0 && rutils.IsInInts(cRole, wantRoles)), 0)
	if req := ctx.GetRequest(); req != nil {
		lng, lat := float64(req.Lng), float64(req.Lat)
		if req.Lng == 0 || req.Lat == 0 {
			lng, lat = userCache.Location.Lon, userCache.Location.Lat
		}
		fs.Add(10005, float32(rutils.EarthDistance(lng, lat, curr.Location.Lon, curr.Location.Lat)/1000.0))
	}

	if userProfile != nil && currProfile != nil {
		// 通过关注的als相关性

		followUser := userProfile.VectorMap["vector_follow_als_user"]
		followFollower := currProfile.VectorMap["vector_follow_als_follower"]
		fs.Add(10006, utils.ArrayMultSum(followUser, followFollower))

		// 通过点击的als相关性
		clickUser := userProfile.VectorMap["vector_click_als_user"]
		clickReceived := currProfile.VectorMap["vector_click_als_received"]
		fs.Add(10007, utils.ArrayMultSum(clickUser, clickReceived))
	}

	return fs
}

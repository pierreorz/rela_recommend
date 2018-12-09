package service

import (
	"math"
	models "rela_recommend/models/pika"
	"rela_recommend/utils"
	"strings"
	"time"
)

var maxFloat64 = math.MaxFloat64
var minFloat64 = -math.MaxFloat64

var userRatio = [][]float64{
	{minFloat64, 62.0},
	{62.0, 77.0},
	{77.0, 85.0},
	{85.0, 92.0},
	{92.0, maxFloat64},
}

var receiverRatio = [][]float64{
	{minFloat64, 58.0},
	{58.0, 67.0},
	{67.0, 75.0},
	{75.0, 83.0},
	{83.0, 85.0},
	{85.0, 92.0},
	{92.0, maxFloat64},
}

var distanceBkt = [][]float64{
	{minFloat64, 9.3591},
	{9.3591, 59.2944},
	{59.2944, 199.8592},
	{199.8592, 623.9141},
	{623.9141, 1043.9025},
	{1043.9025, 1453.8044},
	{1454.8044, 2218.8663},
	{2218.8663, 11522.7978},
	{11522.7978, 12452.5373},
	{12452.5373, maxFloat64},
}

var userCreateDaysBkt = [][]float64{
	{minFloat64, 16.0},
	{16.0, 94.0},
	{94.0, 216.0},
	{216.0, 364.0},
	{364.0, 591.0},
	{591.0, 752.0},
	{752.0, 917.0},
	{917.0, 1078.0},
	{1078.0, 1250.0},
	{1250.0, maxFloat64},
}

var receiverCreateDaysBkt = [][]float64{
	{minFloat64, 106.0},
	{106.0, 243.0},
	{243.0, 408.0},
	{408.0, 629.0},
	{629.0, 770.0},
	{770.0, 907.0},
	{907.0, 1054.0},
	{1054.0, 1138.0},
	{1138.0, 1341.0},
	{1341.0, maxFloat64},
}

var userImageCountBkt = [][]float64{
	{minFloat64, 1.0},
	{1.0, 2.0},
	{2.0, 3.0},
	{3.0, 4.0},
	{4.0, 5.0},
	{5.0, 6.0},
	{6.0, 7.0},
	{7.0, 8.0},
	{8.0, 9.0},
	{9.0, 10.0},
	{10.0, maxFloat64},
}

var receiverImageCountBkt = [][]float64{
	{minFloat64, 2.0},
	{2.0, 3.0},
	{3.0, 4.0},
	{4.0, 5.0},
	{5.0, 7.0},
	{7.0, 10.0},
	{10.0, maxFloat64},
}

var activeScores = []float64{0.0, 0.8, 0.9, 1.0, 1.2, 1.5}

var userAgeRange = [][]float64{
	{minFloat64, 19.0},
	{19.0, 21.0},
	{21.0, 24.0},
	{24.0, 28.0},
	{28.0, maxFloat64},
}

var userHeightRange = [][]float64{
	{minFloat64, 158.0},
	{158.0, 161.0},
	{161.0, 165.0},
	{165.0, 168.0},
	{168.0, maxFloat64},
}

var userWeightRange = [][]float64{
	{minFloat64, 45.0},
	{45.0, 48.0},
	{48.0, 52.0},
	{52.0, 60.0},
	{60.0, maxFloat64},
}

var receiverAgeRange = [][]float64{
	{minFloat64, 19.0},
	{19.0, 21.0},
	{21.0, 23.0},
	{23.0, 26.0},
	{26.0, maxFloat64},
}

var receiverHeightRange = [][]float64{
	{minFloat64, 160.0},
	{160.0, 162.0},
	{162.0, 165.0},
	{165.0, 168.0},
	{168.0, maxFloat64},
}

var receiverWeightRange = [][]float64{
	{minFloat64, 45.0},
	{45.0, 48.0},
	{48.0, 52.0},
	{52.0, 58.0},
	{58.0, maxFloat64},
}

var userMomentsCountRange = [][]float64{
	{minFloat64, 3.0},
	{3.0, 6.0},
	{6.0, 12.0},
	{12.0, 20.0},
	{20.0, 33.0},
	{33.0, 53.0},
	{53.0, 89.0},
	{89.0, 173.0},
	{173.0, maxFloat64},
}

var receiverMomentsCountRange = [][]float64{
	{minFloat64, 1.0},
	{1.0, 3.0},
	{3.0, 7.0},
	{7.0, 12.0},
	{12.0, 21.0},
	{21.0, 34.0},
	{34.0, 54.0},
	{54.0, 90.0},
	{90.0, 168.0},
	{168.0, maxFloat64},
}

func UserRow(user *models.UserProfile, receiver *models.UserProfile) []float32 {
	arr := make([]float32, 0)
	for i := 0; i <= 7; i++ {
		jMax := 7
		if i >= 6 {
			jMax = 5
		}
		for j := 0; j <= jMax; j++ {
			arr = append(arr, getUserReceiverRoleMatch(user.RoleName, receiver.RoleName, i, j))
		}
	}

	for i := -1; i <= 7; i++ {
		for j := -1; j <= 7; j++ {
			arr = append(arr, getUserReceiverAffection(user.Affection, receiver.Affection, i, j))
		}
	}

	userMatchReceiver := IsMatch(user.WantRole, receiver.RoleName)
	receiverMatchUser := IsMatch(receiver.WantRole, user.RoleName)
	var res = userMatchReceiver + receiverMatchUser
	arr = append(arr, propEqualString(res, "00"))
	arr = append(arr, propEqualString(res, "01"))
	arr = append(arr, propEqualString(res, "10"))
	arr = append(arr, propEqualString(res, "11"))

	for _, ar := range userAgeRange {
		arr = append(arr, propBkt(ar[0], ar[1], user.Age))
	}
	for _, ar := range userHeightRange {
		arr = append(arr, propBkt(ar[0], ar[1], user.Height))
	}
	for _, ar := range userWeightRange {
		arr = append(arr, propBkt(ar[0], ar[1], user.Weight))
	}

	for _, ar := range receiverAgeRange {
		arr = append(arr, propBkt(ar[0], ar[1], receiver.Age))
	}
	for _, ar := range receiverHeightRange {
		arr = append(arr, propBkt(ar[0], ar[1], receiver.Height))
	}
	for _, ar := range receiverWeightRange {
		arr = append(arr, propBkt(ar[0], ar[1], receiver.Weight))
	}

	for _, ar := range userMomentsCountRange {
		arr = append(arr, propBkt(ar[0], ar[1], user.MomentsCount))
	}
	for _, ar := range receiverMomentsCountRange {
		arr = append(arr, propBkt(ar[0], ar[1], receiver.MomentsCount))
	}

	for _, ar := range userRatio {
		arr = append(arr, propBkt(ar[0], ar[1], user.Ratio))
	}
	for _, ar := range receiverRatio {
		arr = append(arr, propBkt(ar[0], ar[1], receiver.Ratio))
	}

	var distance = calculateDistance(user.Location.Lon, user.Location.Lat, receiver.Location.Lon, receiver.Location.Lat) / 1000
	for _, ar := range distanceBkt {
		arr = append(arr, propFloatBkt(ar[0], ar[1], distance))
	}

	for _, ar := range userCreateDaysBkt {
		arr = append(arr, propBkt(ar[0], ar[1], calculateCreateDays(user.CreateTime)))
	}
	for _, ar := range receiverCreateDaysBkt {
		arr = append(arr, propBkt(ar[0], ar[1], calculateCreateDays(receiver.CreateTime)))
	}

	var user_image_count = user.UserImageCount + user.NewImageCount
	for _, ar := range userImageCountBkt {
		arr = append(arr, propBkt(ar[0], ar[1], user_image_count))
	}
	
	var receiver_image_count = receiver.UserImageCount + receiver.NewImageCount
	for _, ar := range receiverImageCountBkt {
		arr = append(arr, propBkt(ar[0], ar[1], receiver_image_count))
	}

	arr = append(arr, checkVip(user.IsVip, 0))
	arr = append(arr, checkVip(user.IsVip, 1))

	for i := 0; i <= 7; i++ {
		var ap float32 = 0.0
		if strings.Contains(user.RoleName, utils.GetString(i)) {
			ap = 1.0
		}
		arr = append(arr, ap)
	}

	for i := -1; i <= 7; i++ {
		var ap float32 = 0.0
		if user.Affection == i {
			ap = 1
		}
		arr = append(arr, ap)
	}

	for i := 0; i <= 11; i++ {
		var ap float32 = 0.0
		if user.Horoscope == utils.GetString(i) {
			ap = 1
		}
		arr = append(arr, ap)
	}

	arr = append(arr, checkVip(receiver.IsVip, 0))
	arr = append(arr, checkVip(receiver.IsVip, 1))

	for i := 0; i <= 7; i++ {
		var ap float32 = 0.0
		if strings.Contains(receiver.RoleName, utils.GetString(i)) {
			ap = 1
		}
		arr = append(arr, ap)
	}

	for i := -1; i <= 7; i++ {
		var ap float32 = 0.0
		if receiver.Affection == i {
			ap = 1
		}
		arr = append(arr, ap)
	}

	for i := 0; i <= 11; i++ {
		var ap float32 = 0.0
		if receiver.Horoscope == utils.GetString(i) {
			ap = 1
		}
		arr = append(arr, ap)
	}

	var t = time.Unix(receiver.LastUpdateTime, 0)

	for _, e := range activeScores {
		arr = append(arr, propEqualFloat(activeScore(t), e))
	}

	var now = time.Now()
	for i := 1; i <= 7; i++ {
		arr = append(arr, propEqual(createWeek(now), i))
	}
	for i := 1; i <= 23; i++ {
		arr = append(arr, propEqual(createHour(now), i))
	}
	// 90天单身偏好 352-359
	for i := -1; i <= 7; i++ {
		istr := utils.GetString(i)
		if i != receiver.Affection || user.JsonAffeLike == nil { 
			arr = append(arr, 0.0)
		} else {
			arr = append(arr, user.JsonAffeLike[istr])
		}
	}
	// 90天性别偏好 360-369
	for i := -1; i <= 8; i++ {
		istr := utils.GetString(i)
		if istr != receiver.RoleName || user.JsonRoleLike == nil {
			arr = append(arr, 0.0)
		} else {
			arr = append(arr, user.JsonRoleLike[istr])
		}
	}

	return arr
}

func IsMatch(userWantRole string, receiverRole string) string {
	if strings.Contains(userWantRole, receiverRole) {
		return "1"
	}
	return "0"
}

func createHour(createTime time.Time) int {
	return createTime.Hour()
	//var now = time.Now().Unix()
	//return int((now - createTime.Unix()) / 3600)
}

func createWeek(createTime time.Time) int {
	// var weekday = createTime.Weekday().String()
	// if weekday == "Sunday" {
	// 	return 0
	// } else if weekday == "Monday" {
	// 	return 1
	// } else if weekday == "Tuesday" {
	// 	return 2
	// } else if weekday == "Wednesday" {
	// 	return 3
	// } else if weekday == "Thursday" {
	// 	return 4
	// } else if weekday == "Friday" {
	// 	return 5
	// } else if weekday == "Saturday" {
	// 	return 6
	// }
	// return 0
	return int(createTime.Weekday()) + 1
	//var now = time.Now().Unix()
	//return int((now - createTime.Unix()) / (3600 * 24 * 7))
}

func propEqual(src int, dest int) float32 {
	if src == dest {
		return 1
	}
	return 0
}

func propEqualFloat(src float64, desc float64) float32 {
	if src == desc {
		return 1
	}
	return 0
}

func activeScore(createTime time.Time) float64 {
	now := time.Now().Unix()
	days := (now - createTime.Unix()) / (3600 * 24)
	if days <= 1 {
		return 1.5
	} else if days > 1 && days <= 3 {
		return 1.2
	} else if days > 3 && days <= 7 {
		return 1
	} else if days > 7 && days <= 14 {
		return 0.9
	} else if days > 14 && days <= 30 {
		return 0.8
	}
	return 0
}

func checkVip(isVip int, point int) float32 {
	if isVip == point {
		return 1
	}
	return 0
}

func calculateCreateDays(t time.Time) int {
	var now = time.Now().Unix()
	return int((now - t.Unix()) / (3600 * 24))
}

func calculateDistance(lng1, lat1, lng2, lat2 float64) float64 {
	radius := 6378137.0 // 6378137
	rad := math.Pi / 180.0

	lng1 = lng1 * rad
	lat1 = lat1 * rad

	lng2 = lng2 * rad
	lat2 = lat2 * rad

	theta := lng2 - lng1
	dist := math.Acos(math.Sin(lat1)*math.Sin(lat2) + math.Cos(lat1)*math.Cos(lat2)*math.Cos(theta))
	return dist * radius
}

func propFloatBkt(start float64, end float64, prop float64) float32 {
	var fProp = utils.GetFloat64(prop)
	if start <= fProp && end >= fProp {
		return 1
	}
	return 0
}

func propEqualString(src string, dest string) float32 {
	if src == dest {
		return 1
	}
	return 0
}

func propBkt(start float64, end float64, prop int) float32 {
	var fProp = utils.GetFloat64(prop)
	if start <= fProp && end >= fProp {
		return 1
	}
	return 0
}

func getUserReceiverRoleMatch(from string, to string, i int, j int) float32 {
	if from == utils.GetString(i) && to == utils.GetString(j) {
		return 1
	}
	return 0
}

func getUserReceiverAffection(from int, to int, i int, j int) float32 {
	if from == i && to == j {
		return 1
	}
	return 0
}

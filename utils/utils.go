package utils

import (
	"bytes"
	"crypto/md5"
	crand "crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gansidui/geohash"
)

const MAX_INT64 = 9223372036854775807

const (
	OffTime       = "ER2_L2#EG2#E4"
	RequestErr    = "ER2_L2#EG2#E5"
	RecallOffTime = "L7#EG6#E21"
	RecallOwn     = "L7#EG6#E22"
)

func FormatKeyInt64(str string, i int64) string {
	return fmt.Sprintf("%s%d", str, i)
}

func FormatKeyInt64s(keyFormater string, ids []int64) []string {
	dataLen := len(ids)
	keys := make([]string, dataLen)
	for i, id := range ids {
		keys[i] = fmt.Sprintf(keyFormater, id)
	}
	return keys
}

func ConvertExpId(expId string, recall_expId string) string {
	result := ""
	expArr := strings.Split(expId, "_")
	if len(expArr) == 2 {
		return expArr[0] + "_" + recall_expId + "_" + expArr[1]
	}
	return result
}

func EarthDistance(lng1, lat1, lng2, lat2 float64) float64 {
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

func FormatDistance(lng1, lat1, lng2, lat2 float64) string {
	distance := EarthDistance(lng1, lat1, lng2, lat2)
	var distanceStr string
	if distance > 1000*10 {
		//distance+=dfNotDot.format(s)+" km";
		distanceStr = fmt.Sprintf("%.0f km", distance/1000)

	} else if distance > 1000 {
		//distance+=dfWithDot.format(s)+" km";
		distanceStr = fmt.Sprintf("%.2f km", distance/1000)
	} else {
		// distanceStr = fmt.Sprintf("%.0f m", distance)
		//4.7.5版本小于1000m的距离显示附近
		distanceStr = "< 1 km"
	}
	return distanceStr
}

const X_REGION = "x-region"

func FormatDistanceByLanguage(lng1, lat1, lng2, lat2 float64, language string) string {
	var distanceStr string

	if (lng1 == float64(0) && lat1 == float64(0)) || (lng2 == float64(0) && lat2 == float64(0)) {
		//switch language {
		//case codec.LANGUAGE_CHS:
		//	distanceStr = Chs[X_REGION]
		//case codec.LANGUAGE_CHT:
		//	distanceStr = Cht[X_REGION]
		//case codec.LANGUAGE_CHT_TW:
		//	distanceStr = Cht_tw[X_REGION]
		//case codec.LANGUAGE_EN:
		//	distanceStr = En[X_REGION]
		//case codec.LANGUAGE_ES:
		//	distanceStr = Es[X_REGION]
		//case codec.LANGUAGE_FR:
		//	distanceStr = Fr[X_REGION]
		//case codec.LANGUAGE_TH:
		//	distanceStr = Th[X_REGION]
		//case codec.LANGUAGE_JP:
		//	distanceStr = Jp[X_REGION]
		//default:
		//	distanceStr = Chs[X_REGION]
		//}

		distanceStr = ""

		return distanceStr
	}

	distance := EarthDistance(lng1, lat1, lng2, lat2)

	if distance > 1000*10 {
		//distance+=dfNotDot.format(s)+" km";
		distanceStr = fmt.Sprintf("%.0f km", distance/1000)
	} else if distance >= 1000 {
		//distance+=dfWithDot.format(s)+" km";
		distanceStr = fmt.Sprintf("%.2f km", distance/1000)
	} else if distance > 100 {
		distanceStr = fmt.Sprintf("%d m", int(distance))
	} else {
		// distanceStr = fmt.Sprintf("%.0f m", distance)
		//4.7.5版本小于1000m的距离显示附近
		distanceStr = "< 100 m"
	}
	return distanceStr
}

//获取
func GetGeoHash(lng float64, lat float64, precision int) string {
	hash, _ := geohash.Encode(lat, lng, precision)
	return hash
}

func MaxFloatMap(in map[int64]float64) (key int64, value float64) {
	for k, v := range in {
		if v >= value {
			key = k
			value = v
		}
	}
	return
}

//返回排序结果
func SortMapByValue(in map[int64]float64) []int64 {
	//在go语句中，slice，map，channel这三种类型是拷贝内存地址的，其他的都是拷贝赋值 所以必须在此复制
	out := make([]int64, 0)
	length := len(in)
	for i := 0; i < length; i++ {
		key, _ := MaxFloatMap(in)
		out = append(out, key)
		delete(in, key)
	}
	return out
}

/*
MD5加密
*/
func Md5Sum(data []byte) string {
	return hex.EncodeToString(byte16ToBytes(md5.Sum(data)))
}

func Bytes2Int32(data []byte) int32 {
	var x int32
	b_buf := bytes.NewBuffer(data) // 取最后4字节
	err := binary.Read(b_buf, binary.BigEndian, &x)
	if err != nil {
		fmt.Printf("err %s\n", err)
	}
	return x
}

func Md5Bytes(data []byte) []byte {
	hash := md5.New()
	hash.Write(data)
	resByte := hash.Sum(nil)
	return resByte
}

func Md5Uint32(data []byte) uint32 {
	bys := Md5Bytes(data)
	return binary.BigEndian.Uint32(bys)
}

func Md5Uint64(data []byte) uint64 {
	bys := Md5Bytes(data)
	return binary.BigEndian.Uint64(bys)
}

func Md5Sum32(data []byte) int32 {
	hash := md5.New()
	hash.Write(data)
	resByte := hash.Sum(nil)

	return Bytes2Int32(resByte[12:])
}

//[16]byte to []byte
func byte16ToBytes(in [16]byte) []byte {
	tmp := make([]byte, 16)
	for _, value := range in {
		tmp = append(tmp, value)
	}
	return tmp[16:]
}

func UniqueId() string {
	b := make([]byte, 48)
	if _, err := io.ReadFull(crand.Reader, b); err != nil {
		return ""
	}
	return Md5Sum(b)
}

const httpRegexp = "((http[s]{0,1}|ftp)://[a-zA-Z0-9\\.\\-]+\\.([a-zA-Z]{2,4})(:\\d+)?(/[a-zA-Z0-9\\.\\-~!@#$%^&*+?:_/=<>]*)?)|(www.[a-zA-Z0-9\\.\\-]+\\.([a-zA-Z]{2,4})(:\\d+)?(/[a-zA-Z0-9\\.\\-~!@#$%^&*+?:_/=<>]*)?)"

/**
 * 过滤文中有没有Url.如果有，返回Url路径...
 * @return
 */
func DiscoverUrl(txt string) []string {
	r, _ := regexp.Compile(httpRegexp)
	return r.FindAllString(txt, -1)
}

func Rand(start, end int) int {
	if start >= end {
		return 0
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return start + r.Intn(end-start)
}

func RandomNums(num, limit int) []int {
	var ids = make([]int, 0)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	ids = r.Perm(num)

	if len(ids) < limit {
		return ids
	}

	return ids[:limit]
}

func StartWith(str, subStr string) bool {
	return strings.Index(str, subStr) == 0
}

func Contains(obj interface{}, target interface{}) (bool, error) {
	targetValue := reflect.ValueOf(target)
	switch reflect.TypeOf(target).Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < targetValue.Len(); i++ {
			if targetValue.Index(i).Interface() == obj {
				return true, nil
			}
		}
	case reflect.Map:
		if targetValue.MapIndex(reflect.ValueOf(obj)).IsValid() {
			return true, nil
		}
	}

	return false, errors.New("not in array")
}

func ContainsInt64(arr []int64, target int64) bool {
	for _, element := range arr {
		if element == target {
			return true
		}
	}
	return false
}

func GetStructByJson(data []byte, i interface{}) error {
	return json.Unmarshal(data, i)
}

func GetJsonByStruct(i interface{}) ([]byte, error) {
	return json.Marshal(i)
}

func FormatLongStrToArr(convertStr string) []int64 {
	strArr := strings.Split(convertStr, ",")
	ret := make([]int64, 0)
	for _, intStr := range strArr {
		out, err := strconv.ParseInt(intStr, 10, 64)
		if err != nil {
			continue
		}
		ret = append(ret, out)
	}
	return ret
}

func FormatStrToArrString(convertStr, sep string) []string {
	return strings.Split(convertStr, sep)
}

func Remove(slice []int64, elems int64) []int64 {
	for i := range slice {
		if slice[i] == elems {
			slice = append(slice[:i], slice[i+1:]...)
			return slice
		}
	}
	return slice
}

func Removes(slice []int64, elems []int64) []int64 {
	for _, elem := range elems {
		slice = Remove(slice, elem)
	}
	return slice
}

func GetLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", nil
}

func FormartHttpToHttps(url string) string {
	return strings.Replace(url, "http://", "https://", -1)
}

func AddThumbnail(urlstr string) string {
	URL, err := url.Parse(urlstr)
	if err != nil {
		return urlstr
	}
	q := URL.Query()
	if q.Encode() == "" {
		return urlstr + "?imageView2/2/w/75/h/75/format/webp"
	} else {
		return urlstr + "&imageView2/2/w/75/h/75/format/webp"
	}
}

func Int64ToBytes(i int64) []byte {
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(i))
	return buf
}

func BytesToInt64(buf []byte) int64 {
	return int64(binary.BigEndian.Uint64(buf))
}

// 0 其他， 1 IOS， 2 Android
func GetPlatform(ua string) int {
	ua = strings.ToLower(ua)
	if strings.Contains(ua, "ios") {
		return 1
	} else if strings.Contains(ua, "android") {
		return 2
	}
	return 0
}

func GetPlatformName(ua string) string {
	ua = strings.ToLower(ua)
	if strings.Contains(ua, "ios") {
		return "ios"
	} else if strings.Contains(ua, "android") {
		return "android"
	}
	return "other"
}

func UaAnalysis(ua string) (string, string, string, string, string) {
	osType := ""
	brand := ""
	modelType := ""
	netType := ""
	language := ""
	ua = strings.ToLower(ua)
	if GetPlatformName(ua) == "android" {
		if uaArr := strings.Split(strings.ReplaceAll(ua, ")", ""), "("); len(uaArr) == 2 {
			if itemArr := strings.Split(uaArr[1], ";"); len(itemArr) == 5 {
				modelType = strings.ReplaceAll(itemArr[0], " ", "")
				osType = strings.ReplaceAll(itemArr[1], " ", "")
				netType = strings.ReplaceAll(itemArr[3], " ", "")
				brand = strings.ReplaceAll(itemArr[4], " ", "")
				language = strings.ReplaceAll(itemArr[2], " ", "")
			}
		}
	} else if GetPlatformName(ua) == "ios" {
		if uaArr := strings.Split(strings.ReplaceAll(ua, ")", ""), "("); len(uaArr) == 2 {
			if itemArr := strings.Split(uaArr[1], ";"); len(itemArr) == 4 {
				modelType = strings.ReplaceAll(itemArr[0], " ", "")
				osType = strings.ReplaceAll(itemArr[1], " ", "")
				netType = strings.ReplaceAll(itemArr[3], " ", "")
				brand = "apple"
				language = strings.ReplaceAll(itemArr[2], " ", "")
			}
		}
	}
	return osType, brand, modelType, netType, language

}

var versionRe = regexp.MustCompile(`/(\d{1,2})\.(\d{1,2})(?:\.(\d{1,2}))?`)

// 获取版本信息：5.3.6 返回 5.3.6; 5.4 返回 5.4.0
func GetVersionString(ua string) string {
	if matched := versionRe.FindAllStringSubmatch(ua, -1); len(matched) > 0 {
		var vs = []string{}
		for _, s := range matched[0][1:] {
			vs = append(vs, IfElseString(s != "", s, "0"))
		}
		return strings.Join(vs, ".")
	}
	return ""
}

// 获取版本信息：5.3.6 返回 50306; 5.4 返回 50400
func GetVersion(ua string) int {
	var version = 0
	// theL/5.20.4  matched [[/5.20.4 5 20 4]]
	if vs := strings.Split(GetVersionString(ua), "."); len(vs) > 0 {
		// fmt.Printf("\t%s  match: %s \n", ua, strings.Join(vs, ","))
		for _, s := range vs {
			if v, err := strconv.Atoi(s); err == nil {
				version = version*100 + v
			}
		}
	}
	return version
}

// int是否在int数组内
func IsInInts(v int, vs []int) bool {
	for _, vv := range vs {
		if v == vv {
			return true
		}
	}
	return false
}

// 分隔数组
func SplitList(list []interface{}, partLen int) [][]interface{} {
	arrs := make([][]interface{}, partLen)
	dataLen, endLen := len(list)/partLen, len(list)%partLen
	for i := 0; i < partLen; i++ {
		startIndex, endIndex := i*dataLen, (i+1)*dataLen
		if i == partLen-1 {
			endIndex += endLen
		}
		arrs = append(arrs, list[startIndex:endIndex])
	}
	return arrs
}

type partIndex struct {
	Start int
	End   int
}

// 分隔数组
func SplitPartIndex(dataLen, partLen int) []partIndex {
	indexs := make([]partIndex, partLen)
	dataLen, endLen := dataLen/partLen, dataLen%partLen
	for i := 0; i < partLen; i++ {
		startIndex, endIndex := i*dataLen, (i+1)*dataLen
		if i == partLen-1 {
			endIndex += endLen
		}
		indexs[i] = partIndex{Start: startIndex, End: endIndex}
	}
	return indexs
}

// 返回第一个不为空的字符串
func CoalesceString(strs ...string) string {
	for _, str := range strs {
		if len(str) > 0 {
			return str
		}
	}
	return ""
}

func CoalesceInt(ints ...int) int {
	for _, i := range ints {
		if i != 0 {
			return i
		}
	}
	return 0
}

func IfElse(b bool, trueValue float64, falseValue float64) float64 {
	if b {
		return trueValue
	} else {
		return falseValue
	}
}

func IfElseString(b bool, trueValue string, falseValue string) string {
	if b {
		return trueValue
	} else {
		return falseValue
	}
}

// 字符串中是否存在某个字符，任意一个返回true
func StringContains(str string, words []string) bool {
	for _, word := range words {
		if strings.Contains(str, word) {
			return true
		}
	}
	return false
}

// 多char分割字符串
func Splits(s string, chars string) []string {
	var charMap = map[rune]bool{}
	for _, char := range chars {
		charMap[char] = true
	}
	return strings.FieldsFunc(s, func(c rune) bool {
		_, ok := charMap[c]
		return ok
	})
}

func JoinInt64s(int64s []int64, spliter string) string {
	strs := make([]string, len(int64s))
	for k, v := range int64s {
		strs[k] = fmt.Sprintf("%d", v)
	}
	return strings.Join(strs, ",")
}

func MergeMapStringFloat64(maps ...map[string]float64) map[string]float64 {
	var res = map[string]float64{}
	if len(maps) > 0 {
		res = maps[0]
		for _, mp := range maps[1:] {
			for mpk, mpv := range mp {
				res[mpk] = res[mpk] + mpv
			}
		}
	}
	return res
}

func GaussDecay(x, center, offset, scale float64) float64 {
	// center 高斯函数中心点
	// offset 从 center 为中心，为他设置一个偏移量offset覆盖一个范围，在此范围内所有的概率也都是和 center 一样满分
	// scale 衰减到0.5时，offset到当前x的差值
	if math.Abs(x-center) <= offset {
		x = 0.
	} else {
		x = x - center - offset
	}
	sigma := -math.Pow(scale, 2) / (2 * math.Log(0.5))
	return math.Pow(math.E, -math.Pow(x, 2)/(2*sigma))
}

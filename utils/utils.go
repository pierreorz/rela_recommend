package utils

import (
	"crypto/md5"
	crand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"io"
	"encoding/binary"
	"bytes"
	"github.com/gansidui/geohash"
	"net/url"
)

const MAX_INT64 = 9223372036854775807

func FormatKeyInt64(str string, i int64) string {
	return fmt.Sprintf("%s%d", str, i)
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
	b_buf := bytes.NewBuffer(data)  // 取最后4字节
	err := binary.Read(b_buf, binary.BigEndian, &x) 
	if err != nil{
		fmt.Printf("err %s\n", err)
	}
	return x
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
	if start < end {
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
	for i, _ := range slice {
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

// Set 结构
type SetInt64 struct {
	intMap map[int64]int
}

func(self *SetInt64) checkMap(setEmpty bool) {
	if self.intMap == nil || setEmpty {
		self.intMap = make(map[int64]int, 0)
	}
}

func(self *SetInt64) FromArray(vals []int64) {
	self.checkMap(true)
	for _, val := range vals {
		self.intMap[val] = 1
	}
}

func(self *SetInt64) AppendArray(vals []int64) *SetInt64 {
	self.checkMap(false)
	for _, val := range vals {
		self.intMap[val] = 1
	}
	return self
}


func(self *SetInt64) Contains(val int64) bool {
	_, ok := self.intMap[val]
	return ok
}

func(self *SetInt64) ToList() []int64 {
	res := make([]int64, 0)
	for k, _ := range self.intMap {
		res = append(res, k)
	}
	return res
}

func NewSetInt64FromArray(vals []int64) *SetInt64 {
	set := SetInt64{}
	set.FromArray(vals)
	return &set
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
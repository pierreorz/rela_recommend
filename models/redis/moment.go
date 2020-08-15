package redis

import (
	"fmt"
	"rela_recommend/utils"
	"time"

	"rela_recommend/algo"
	"rela_recommend/cache"

	"github.com/gansidui/geohash"
)

type Moments struct {
	Id int64 `gorm:"primary_key;column:id" json:"id"`
	/** 发表文章的用户 **/
	UserId int64 `gorm:"column:user_id" json:"userId,omitempty"`
	/** 日志状态 默认1正常, 0已删除 **/
	Status int8 `gorm:"default:1;column:status" json:"status,omitempty"`
	/** 分享(all, friends, only_me) **/
	ShareTo string `gorm:"column:share_to" json:"shareTo,omitempty"`
	/**创建时间**/
	InsertTime time.Time `gorm:"column:insert_time" json:"insertTime,omitempty"`
	/**最后更新时间**/
	LastUpdateTime time.Time `gorm:"default:NULL;column:last_update_time" json:"last_update_time,omitempty"`
	/** 文章类型(text, image, voice, text_image, text_voice, image_voice) **/
	MomentsType string `gorm:"column:moments_type" json:"momentsType,omitempty"`
	/** 文章的文字内容 **/
	MomentsText string `gorm:"column:moments_text" json:"momentsText,omitempty"`
	/** 缩略图 ... 注意：已经改为存储‘是否@了别人’，如果‘@了别人’，则保存为“1”。。。至于‘缩略图’，可以根据大图来推算出来  **/
	ThumbnailUrl string `gorm:"column:thumbnail_url" json:"thumbnailUrl,omitempty"`
	/** 文章的配图链接 **/
	ImageUrl string `gorm:"column:image_url" json:"imageUrl,omitempty"`
	/** 因为视频日志需要搜索排列，所以只能使用新建列存储，不能放到ext列中, 如果所有日志信息都能放到搜索引擎中就可以避免使用此列 **/
	VoiceUrl string `gorm:"column:voice_url" json:"voiceUrl,omitempty"`
	/** 是否匿名(0表示不匿名,1表示匿名) **/
	Secret int8 `gorm:"default:0;column:secret" json:"secret,omitempty"`
	/** 目前没有什么用  **/
	//WinkCount int `gorm:"column:wink_count" json:"winkCount,omitempty"`
	/** 版本号 目前没有什么用  **/
	//Version int `gorm:"column:version" json:"version,omitempty"`
	/** 额外信息  **/
	Ext string `gorm:"column:-" json:"-,omitempty"`
	/** 额外信息  **/
	MomentsExt MomentsExt `gorm:"column:ext" json:"ext,omitempty"`
}

type MomentsExt struct {
	ThemeClass      string `json:"themeClass,omitempty"`
	ThemeReplyClass string `json:"themeReplyClass,omitempty"`
	AdUrl           string `json:"adUrl,omitempty"`
	AdType          string `json:"adType,omitempty"`
	AppSchemeUrl    string `json:"appSchemeUrl,omitempty"`
	VideoWebp       string `json:"videoWebp,omitempty"`
	VideoColor      string `json:"videoColor,omitempty"`
	VideoType       string `json:"videoType,omitempty"`    // 4.7.3视频新增类型 PGC 官方 UGC 个人
	IsCoverImage    bool   `json:"isCoverImage,omitempty"` // 4.9.1封面图
	IsLandscape     int    `json:"isLandscape,omitempty"`  // 横屏
	SyncMainPage    bool   `json:"syncMainPage,omitempty"` //是否同步到主页
	AtUserList      string `json:"atUserList,omitempty"`   //提及用户列表
	TagList         string `json:"tagList,omitempty"`      //标签组
	IsFive          int    `json:"isFive,omitempty"`       //5.0版本此值为1
	Reason          string `json:"reason,omitempty"`       //推荐网页的理由
}

type MomentsExtend struct {
	MomentsId   int64  `gorm:"column:moments_id" json:"momentsId"`               //日志的 id
	ImgLen      int    `gorm:"column:img_len" json:"imgLen,omitempty"`           //有图片的日志，大图的大小（字节数），无图就写0
	AndroidFlag int8   `gorm:"column:android_flag" json:"androidFlag,omitempty"` //是否安卓系统发的日志(1表示安卓,0表示苹果)
	MobileOs    string `gorm:"column:mobile_os" json:"mobileOs,omitempty"`       //完整的系统版本号，例如：Android 4.4.1
	// chs cht en
	MomentsLang string `gorm:"column:moments_lang" json:"momentsLang,omitempty"` //日志语言类型
	Language    string `gorm:"column:language" json:"language,omitempty"`        //日志语言，手机端语言
	//add: 位置信息
	MomentsType string    `gorm:"column:moments_type" json:"momentsType,omitempty"` //日志类型
	UserId      int64     `gorm:"column:user_id" json:"userId,omitempty"`           //用户信息
	InsertTime  time.Time `gorm:"column:insert_time" json:"insertTime"`
	Lng         float64   `gorm:"column:lng" json:"lng"`
	Lat         float64   `gorm:"column:lat" json:"lat"`
	/** 日志状态 默认1正常, 0已删除 **/
	Status              int8   `gorm:"column:status" json:"status,omitempty"`
	VoiceUrl            string `gorm:"column:voice_url" json:"voiceUrl,omitempty"`                         //当类型为音乐时： 为日志的语音，音乐链接  , 视频日志： 视频链接，     直播日志： liveId
	VoiceName           string `gorm:"column:voice_name" json:"voiceName,omitempty"`                       //歌曲名
	VoiceAlbum          string `gorm:"column:voice_album" json:"voiceAlbum,omitempty"`                     //所属专辑
	VoiceAuthor         string `gorm:"column:voice_author" json:"voiceAuthor,omitempty"`                   //演唱者
	VoiceTime           int    `gorm:"column:voice_time" json:"voiceTime,omitempty"`                       //时长
	DeleteBySelfFlag    int    `gorm:"column:delete_by_self_flag" json:"deleteBySelfFlag,omitempty"`       //是否自己删除的日志(1表示是,0表示否)
	SongId              int64  `gorm:"column:song_id" json:"songId,omitempty"`                             //歌曲id
	VoiceAlbumLogoSmall string `gorm:"column:voice_album_logo_small" json:"voiceAlbumLogoSmall,omitempty"` //专辑图片地址 100X100
	VoiceAlbumLogo      string `gorm:"column:voice_album_logo" json:"voiceAlbumLogo,omitempty"`            //专辑图片地址 444X444
	ToUrl               string `gorm:"column:to_url" json:"toUrl,omitempty"`                               //虾米url
	Pixel               string `gorm:"column:pixel" json:"pixel,omitempty"`                                //像素
	LinkTopicFlag       int8   `gorm:"column:link_topic_flag" json:"linkTopicFlag,omitempty"`              //是否关联话题
}

type MomentsProfile struct {
	LikeCnt          int      `json:"likeCnt,omitempty"`
	TextCnt          int      `json:"textCnt,omitempty"`
	MomentsTextWords []string `json:"momentsTextWords,omitempty"`
}

type MomentOfflineProfile struct {
	Id              int64     `json:"moment_id"`
	MomentEmbedding []float32 `json:"moment_embedding"`
}
type MomentsAndExtend struct {
	Moments        *Moments        `gorm:"column:moments" json:"moments,omitempty"`
	MomentsExtend  *MomentsExtend  `gorm:"column:moments_extend" json:"momentsExtend,omitempty"`
	MomentsProfile *MomentsProfile `gorm:"column:moments_profile" json:"momentsProfile,omitempty"`
}

type MomentUserProfile struct {
	UserID       int64              `json:"user_id"`
	UserEmbedding       []float32  `json:"user_embedding"`
	UserPref             []string `json:"user_pref,omitempty"`

}

type MomentCacheModule struct {
	CachePikaModule
}

func NewMomentCacheModule(ctx algo.IContext, cache *cache.Cache, store *cache.Cache) *MomentCacheModule {
	return &MomentCacheModule{CachePikaModule{ctx: ctx, cache: *cache, store: *store}}
}

// 从缓存中获取以逗号分割的字符串，并转化成int64. 如 keys11  1,2,3,4,5
func (self *MomentCacheModule) GetInt64ListFromGeohash(lat float32, lng float32, len int, keyFormatter string) ([]int64, error) {
	geohash, _ := geohash.Encode(float64(lat), float64(lng), len)
	res, err := self.GetSet(fmt.Sprintf(keyFormatter, geohash), 24*60*60, 1*60*60)
	if err == nil {
		return utils.GetInt64s(utils.GetString(res)), nil
	}
	return nil, err
}

// 读取用户embedding特征
func (self *UserCacheModule) QueryMomentUserProfileByIds(ids []int64) ([]MomentUserProfile, error) {
	keyFormatter := "moment_user_profile:%d"
	ress, err := self.MGetStructs(MomentUserProfile{}, ids, keyFormatter, 24*60*60, 60*60*1)
	objs := ress.Interface().([]MomentUserProfile)
	return objs, err
}

// 获取当前用户和用户列表Map
func (this *UserCacheModule) QueryMomentUserProfileByUserAndUsersMap(userId int64, userIds []int64) (*MomentUserProfile, map[int64]*MomentUserProfile, error) {
	allIds := append(userIds, userId)
	users, err := this.QueryMomentUserProfileByIds(allIds)
	var resUser *MomentUserProfile
	var resUsersMap = make(map[int64]*MomentUserProfile, 0)
	if err == nil {
		for i, user := range users {
			if user.UserID == userId {
				resUser = &users[i]
			} else {
				resUsersMap[user.UserID] = &users[i]
			}
		}
	}
	return resUser, resUsersMap, err
}

//读取日志画像特征
func (self *MomentCacheModule) QueryMomentOfflineProfileByIds(ids []int64) ([]MomentOfflineProfile, error) {
	keyFormatter := "moment_offline_profile:%d"
	ress, err := self.MGetStructs(MomentOfflineProfile{}, ids, keyFormatter, 24*60*60, 60*60*1)
	objs := ress.Interface().([]MomentOfflineProfile)
	return objs, err
}

// 获取当前用户和用户列表Map
func (this *MomentCacheModule) QueryMomentOfflineProfileByIdsMap(momentIds []int64) (map[int64]*MomentOfflineProfile, error) {
	moments, err := this.QueryMomentOfflineProfileByIds(momentIds)
	var resMomentsMap = make(map[int64]*MomentOfflineProfile, 0)
	if err == nil {
		for i, moment := range moments {
			resMomentsMap[moment.Id] = &moments[i]
		}
	}
	return resMomentsMap, err
}

// 读取直播相关用户画像
func (self *MomentCacheModule) QueryMomentsByIds(ids []int64) ([]MomentsAndExtend, error) {
	keyFormatter := "friends_moments_moments:%d"
	ress, err := self.MGetStructs(MomentsAndExtend{}, ids, keyFormatter, 24*60*60, 60*60*1)
	objs := ress.Interface().([]MomentsAndExtend)
	return objs, err
}

func (self *MomentCacheModule) QueryMomentsMapByIds(ids []int64) (map[int64]MomentsAndExtend, error) {
	momsMap := map[int64]MomentsAndExtend{}
	moms, err := self.QueryMomentsByIds(ids)
	if err == nil {
		for i, mom := range moms {
			if mom.Moments != nil {
				momsMap[mom.Moments.Id] = moms[i]
			}
		}
	}
	return momsMap, err
}

func (self *MomentCacheModule) GetInt64ListOrDefault(id int64, defaultId int64, keyFormatter string) ([]int64, error) {
	var resInt64s = make([]int64, 0)
	res, err := self.GetSet(fmt.Sprintf(keyFormatter, id), 6*60*60, 1*60*60)
	if err == nil {
		resInt64s = utils.GetInt64s(utils.GetString(res))
	}
	if len(resInt64s) == 0 {
		res, err := self.GetSet(fmt.Sprintf(keyFormatter, defaultId), 6*60*60, 1*60*60)
		if err == nil {
			resInt64s = utils.GetInt64s(utils.GetString(res))
		}
	}
	return resInt64s, err
}

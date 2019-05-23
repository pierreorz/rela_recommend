package redis

import(
	"time"
	"encoding/json"
	"rela_recommend/log"
	"rela_recommend/cache"
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
	Ext string `gorm:"column:ext" json:"-"`
	/** 额外信息  **/
	MomentsExt MomentsExt `gorm:"column:-" json:"ext,omitempty"`
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
	LikeCnt				int			`json:"likeCnt,omitempty"`  
	TextCnt				int			`json:"textCnt,omitempty"` 
	MomentsTextWords 	[]string	`json:"momentsTextWords,omitempty"` 
}

type MomentsAndExtend struct {
	Moments 		*Moments		`gorm:"column:moments" json:"moments,omitempty"`        
	MomentsExtend	*MomentsExtend	`gorm:"column:moments_extend" json:"momentsExtend,omitempty"`        
	MomentsProfile	*MomentsProfile	`gorm:"column:moments_profile" json:"momentsProfile,omitempty"` 
}

type MomentCacheModule struct {
	CachePikaModule
}

func NewMomentCacheModule(cache *cache.Cache, store *cache.Cache) *MomentCacheModule {
	return &MomentCacheModule{CachePikaModule{cache: *cache, store: *store}}
}

// 读取直播相关用户画像
func (self *MomentCacheModule) QueryMomentsByIds(ids []int64) ([]MomentsAndExtend, error) {
	startTime := time.Now()
	keyFormatter := "friends_moments_moments:%d"
	ress, err := self.MGetSet(ids, keyFormatter, 24 * 60 * 60, 60 * 60 * 1)
	startJsonTime := time.Now()
	objs := make([]MomentsAndExtend, 0)
	for i, res := range ress {
		if res != nil {
			var obj MomentsAndExtend
			bs, ok := res.([]byte)
			if ok {
				if err := json.Unmarshal(bs, &obj); err == nil {
					objs = append(objs, obj)
				} else {
					log.Warn(keyFormatter, ids[i], err.Error())
				}
			} else {
				log.Warn(keyFormatter, ids[i], err.Error())
			}
		}
	}
	endTime := time.Now()
	log.Infof("UnmarshalKey:%s,all:%d,notfound:%d,final:%d;total:%.4f,read:%.4f,json:%.4f\n",
		keyFormatter, len(ids), len(ids)-len(objs), len(objs), 
		endTime.Sub(startTime).Seconds(),
		startJsonTime.Sub(startTime).Seconds(), endTime.Sub(startJsonTime).Seconds())
	return objs, err
}

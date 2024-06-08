package models

import (
	"errors"
	"time"

	"rela_recommend/cache"
	"rela_recommend/help"
	"rela_recommend/log"
	"rela_recommend/utils"

	"github.com/jinzhu/gorm"
)

const (
	PREFIX_USER_LOGIN = "login:"
	PREFIX_USER_KEY   = "user:key:"
	EXPIRE_USER_LOGIN = 3600 * 12
	EXPIRE_NOT_FOUND  = 300
)

const (
	USER_LOGIN_ONLINE = 1
)

var ErrNoRows = errors.New("record not found or disable")

type Login struct {
	UserId         int64     `gorm:"column:user_id" json:"userId"`
	Key            string    `gorm:"column:key" json:"key,omitempty"`
	LoginDate      time.Time `gorm:"column:login_date" json:"loginDate,omitempty"`
	Lng            float64   `gorm:"column:lng" json:"lng,omitempty"`
	Lat            float64   `gorm:"column:lat" json:"lat,omitempty"`
	ClientLanguage string    `gorm:"column:cilent_language" json:"clientLanguage,omitempty"`
	DeviceId       string    `gorm:"column:device_id" json:"deviceId,omitempty"`
	AppType        string    `gorm:"column:app_type" json:"appType,omitempty"`
	ClientVersion  string    `gorm:"column:client_version" json:"clientVersion,omitempty"`
	Ua             string    `gorm:"column:ua" json:"ua,omitempty"`
	//DeviceToken    string    `gorm:"column:device_token" json:"deviceToken,omitempty"`
	Online int8   `gorm:"column:online" json:"online,omitempty"`
	Cid    string `gorm:"column:cid" json:"cid,omitempty"`
	//ImageId         int64 `gorm:"column:image_id" json:"imageId,omitempty"`
	WinkPush     int8 `gorm:"column:wink_push" json:"winkPush,omitempty"`
	FollowerPush int8 `gorm:"column:follower_push" json:"followerPush,omitempty"`
	//KeyPush         int8 `gorm:"column:key_push" json:"keyPush,omitempty"`
	MessagePush     int8 `gorm:"column:message_push" json:"messagePush,omitempty"`
	CommentTextPush int8 `gorm:"column:comment_text_push" json:"commentTextPush,omitempty"`
	CommentUserPush int8 `gorm:"column:comment_user_push" json:"commentUserPush,omitempty"`
	CommentWinkPush int8 `gorm:"column:comment_wink_push" json:"commentWinkPush,omitempty"`
	LivePush        int8 `gorm:"column:live_push" json:"livePush,omitempty"`
}

func (this *Login) TableName() string {
	return "app_user_login"
}

type ILoginModule interface {
	QueryByKey(string, *Login) error
	QueryByUserId(int64, *Login) error
	QueryByUserIdWithDB(int64, *Login) error
	QueryDeviceIdByUserId(int64) (string, error)
	DeleteByUserId(int64) error
	DeleteCacheByUserId(int64) error
	UpdateCacheByUserId(int64, *Login) error
	UpdatePositionByUserId(userId int64, lng, lat float64, clientLanguage, deviceId, appType, clientVersion, userAgent string) error
	UpdatePush(*Login) error
	UpdateLogout(*Login) error
	UpdateImageIdByUserId(int64, int64) error
}

type LoginModule struct {
	db   *gorm.DB
	cach cache.Cache
}

func NewLoginModule(db *gorm.DB, cach cache.Cache) ILoginModule {
	return &LoginModule{db: db, cach: cach}
}

func (this *LoginModule) QueryByKey(key string, pLogin *Login) error {
	return this.db.Where("`key` = ?", key).First(pLogin).Error
}

func (this *LoginModule) QueryByUserId(userId int64, pLogin *Login) error {
	if userId <= 0 {
		return ErrNoRows
	}

	key := utils.FormatKeyInt64(PREFIX_USER_LOGIN, userId)
	if err := help.GetStructByCache(this.cach, key, pLogin); err != nil {
		if err := this.db.Where("user_id = ?", userId).First(pLogin).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				//设置-1,防止数据库穿透,目前还有主从不同步的原因,所以缓存时间不能设置太长
				pLogin.UserId = -1
				if err := help.SetExStructByCache(this.cach, key, pLogin, EXPIRE_NOT_FOUND); err != nil {
					log.Error(err.Error())
				}
			}
			return err
		}
		pLogin.LoginDate = pLogin.LoginDate.Add(-8 * 3600 * time.Second)
		if err := help.SetExStructByCache(this.cach, key, pLogin, EXPIRE_USER_LOGIN); err != nil {
			log.Error(err.Error())
		}
	}
	if pLogin.UserId == -1 {
		return ErrNoRows
	}
	return nil
}

func (this *LoginModule) QueryByUserIdWithDB(userId int64, pLogin *Login) error {
	if userId <= 0 {
		return ErrNoRows
	}
	return this.db.Where("user_id = ?", userId).First(pLogin).Error
}

func (this *LoginModule) DeleteByUserId(userId int64) error {
	if err := this.db.Where("user_id = ?", userId).Delete(&Login{}).Error; err != nil {
		return err
	}
	this.DeleteCacheByUserId(userId)
	return nil
}

func (this *LoginModule) UpdatePositionByUserId(userId int64, lng, lat float64, clientLanguage, deviceId, appType, clientVersion, userAgent string) error {
	var data = map[string]interface{}{}
	data["login_date"] = time.Now().Add(8 * 3600 * time.Second)
	data["online"] = USER_LOGIN_ONLINE
	data["lat"] = lat
	data["lng"] = lng
	if clientLanguage != "" {
		data["client_language"] = clientLanguage
	}
	if deviceId != "" {
		data["device_id"] = deviceId
	}
	if appType != "" {
		data["app_type"] = appType
	}
	if clientVersion != "" {
		data["client_version"] = clientVersion
	}
	if userAgent != "" {
		data["ua"] = userAgent
	}

	if err := this.db.Model(&Login{}).Where("user_id = ?", userId).Update(data).Error; err != nil {
		return err
	}
	//DelUserLoginByUserId(userId)
	this.DeleteCacheByUserId(userId)
	return nil
}

func (this *LoginModule) QueryDeviceIdByUserId(userId int64) (string, error) {
	rows, err := this.db.DB().Query("select device_id from app_user_login where user_id = ? limit 1", userId)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	var deviceId []byte
	for rows.Next() {
		err = rows.Scan(&deviceId)
		if err != nil {
			return "", err
		}
	}
	return string(deviceId), nil
}

func (this *LoginModule) DeleteCacheByUserId(userId int64) error {
	key := utils.FormatKeyInt64(PREFIX_USER_LOGIN, userId)
	return this.cach.Del(key)
}

func (this *LoginModule) UpdateCacheByUserId(id int64, pLogin *Login) error {
	key := utils.FormatKeyInt64(PREFIX_USER_LOGIN, id)
	if err := help.SetExStructByCache(this.cach, key, pLogin, EXPIRE_USER_LOGIN); err != nil {
		log.Error(err.Error())
	}
	return nil
}

func (this *LoginModule) UpdatePush(pLogin *Login) error {
	var data = map[string]interface{}{}
	data["wink_push"] = pLogin.WinkPush
	data["follower_push"] = pLogin.FollowerPush
	//data["key_push"] = pLogin.KeyPush
	data["message_push"] = pLogin.MessagePush
	data["comment_text_push"] = pLogin.CommentTextPush
	data["comment_user_push"] = pLogin.CommentUserPush
	data["comment_wink_push"] = pLogin.CommentWinkPush
	data["live_push"] = pLogin.LivePush

	if err := this.db.Model(pLogin).Where("user_id = ?", pLogin.UserId).Update(data).Error; err != nil {
		return err
	}

	var login Login
	this.QueryByUserId(pLogin.UserId, &login)
	login.WinkPush = pLogin.WinkPush
	login.FollowerPush = pLogin.FollowerPush
	login.MessagePush = pLogin.MessagePush
	login.CommentTextPush = pLogin.CommentTextPush
	login.CommentUserPush = pLogin.CommentUserPush
	login.CommentWinkPush = pLogin.CommentWinkPush
	login.LivePush = pLogin.LivePush
	this.UpdateCacheByUserId(pLogin.UserId, &login)
	return nil
}

func (this *LoginModule) UpdateLogout(pLogin *Login) error {
	var data = map[string]interface{}{}
	data["online"] = pLogin.Online
	data["key"] = pLogin.Key

	if err := this.db.Model(*pLogin).Where("user_id = ?", pLogin.UserId).Update(data).Error; err != nil {
		return err
	}

	var login Login
	this.QueryByUserId(pLogin.UserId, &login)

	login.Online = pLogin.Online
	login.Key = pLogin.Key
	this.UpdateCacheByUserId(pLogin.UserId, &login)
	this.cach.Del(utils.FormatKeyInt64(PREFIX_USER_KEY, login.UserId))
	return nil
}

func (this *LoginModule) UpdateImageIdByUserId(userId int64, imageId int64) error {
	var data = map[string]interface{}{}
	data["image_id"] = imageId
	if err := this.db.Model(&Login{}).Where("user_id = ?", userId).Update(data).Error; err != nil {
		return err
	}
	/*
		var login Login
		this.QueryByUserId(userId, &login)
		login.ImageId = imageId
		this.UpdateCacheByUserId(userId, &login)
	*/
	return nil
}

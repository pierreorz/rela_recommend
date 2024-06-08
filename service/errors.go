package service

import (
	"errors"
)

var (
	ErrInvaPara                 = errors.New("invalid param")
	ErrInvaKey                  = errors.New("invalid key")
	ErrUserDisable              = errors.New("disable user")
	ErrUserUnregister           = errors.New("unregister user")
	ErrUserNotFonud             = errors.New("the user have no found")
	ErrUserAlreadyReport        = errors.New("the user have already report")
	ErrReportYourself           = errors.New("ErrReportYourself")
	ErrMomNotFound              = errors.New("the moment have not found")
	ErrComNotFound              = errors.New("the comment have not found")
	ErrComNoPer                 = errors.New("you have no per for operate other comment")
	ErrMomNoPer                 = errors.New("you have no per for operate other moment")
	ErrUserNoPer                = errors.New("you have no per for operate other user")
	ErrTopic                    = errors.New("invalid topic")
	ErrNotVIP                   = errors.New("not vip")
	ErrUserBeenBlock            = errors.New("the user has been Block")
	ErrCommentText              = errors.New("invalid comment Text")
	ErrCommentType              = errors.New("invalid comment Type")
	ErrMomAlreadyPublish        = errors.New("the moments already publish")
	ErrComAlreadyPublish        = errors.New("the comment already publish")
	ErrUserTodayMomnetsLimit    = errors.New("ErrUserTodayMomnetsLimit")
	ErrUserTodayCommentLimit    = errors.New("ErrUserTodayCommentLimit")
	ErrUserTodayWinkLimit       = errors.New("ErrUserTodayWinkLimit")
	ErrMomAlreadyWink           = errors.New("you have already wink for moment")
	ErrUserWinkToHerSelf        = errors.New("you wink to yourself")
	ErrUserAlreadyWink          = errors.New("you have already wink for user")
	ErrWrongWithThemeReply      = errors.New("ErrWrongWithThemeReply")
	ErrComAlreadyReport         = errors.New("you have alredy report this comment")
	ErrRefreshTooFast           = errors.New("ErrRefreshTooFast")
	ErrNoAds                    = errors.New("No Ads")
	ErrSensitiveWords           = errors.New("ErrSensitiveWords")
	ErrMomAlreadyReport         = errors.New("ErrMomAlreadyReport")
	ErrMomAlreadyHide           = errors.New("ErrMomAlreadyHide")
	ErrRequestNotFonud          = errors.New("ErrRequestNotFonud")
	ErrBffOverLimit             = errors.New("ErrBffOverLimit")
	ErrSheAlreadyBind           = errors.New("ErrSheAlreadyBind")
	ErrYouAlreadyBind           = errors.New("ErrYouAlreadyBind")
	ErrCancelBind               = errors.New("ErrCancelBind")
	ErrYouAlreadyBindShe        = errors.New("ErrYouAlreadyBindShe")
	ErrYouAlreadyBffShe         = errors.New("ErrYouAlreadyBffShe")
	ErrRequestTimeOut           = errors.New("ErrRequestTimeOut")
	ErrSheAlreadySendRequest    = errors.New("ErrSheAlreadySendRequest")
	ErrYouAlreadySendRequest    = errors.New("ErrYouAlreadySendRequest")
	ErrHandleRequest            = errors.New("ErrHandleRequest")
	ErrFollowSelf               = errors.New("ErrFollowSelf")
	ErrAlreadyFollow            = errors.New("ErrAlreadyFollow")
	ErrUserTodayFollowLimit     = errors.New("ErrUserTodayFollowLimit")
	ErrUserNotFriend            = errors.New("ErrUserNotFriend")
	ErrUserImageListOverLimit   = errors.New("ErrUserImageListOverLimit")
	ErrPicNotExist              = errors.New("the pic have not exist")
	ErrPicNoPer                 = errors.New("ErrPicNoPer")
	ErrPicPrivacy               = errors.New("ErrPicPrivacy")
	ErrPicMustHaveOne           = errors.New("ErrPicMustHaveOne")
	ErrBeenShield               = errors.New("ErrBeenShield")
	ErrShieldShe                = errors.New("ErrShieldShe")
	ErrWorldListNotVip          = errors.New("ErrWorldListNotVip")
	ErrLoveMyself               = errors.New("ErrLoveMyself")
	ErrNicknameModifyFrequently = errors.New("nickname can not be modified within 30 days")
	ErrNoPic                    = errors.New("no pic")
	ErrNoPerfectInformation     = errors.New("ErrNoPerfectInformation")
	ErrRepeatUserName           = errors.New("ErrRepeatUserName")
	ErrBlcakMyself              = errors.New("ErrBlcakMyself")
	ErrAlreadyBlcak             = errors.New("ErrAlreadyBlcak")
	ErrOfficeAccount            = errors.New("ErrOfficeAccount")
	ErrUnbindCellphone          = errors.New("Unbind Cellphone")
	ErrNoMatchUser              = errors.New("ErrNoMatchUser")
	ErrMatchTodayLimit          = errors.New("ErrMatchTodayLimit")
	ErrApiLimit                 = errors.New("ErrApiLimit")
	ErrNewVSOldPassEqual        = errors.New("ErrNewVSOldPassEqual")
	ErrOldPassNotEqual          = errors.New("ErrOldPassNotEqual")
)

type IAPIError interface {
	error
	GetError() error
	GetMessage() string
	GetLanguage() string
}

type APIError struct {
	err      error
	language string
	message  string
}

func (api *APIError) Error() string {
	return api.err.Error()
}

func (api *APIError) GetError() error {
	return api.err
}

func (api *APIError) GetMessage() string {
	return api.message
}

func (api *APIError) GetLanguage() string {
	return api.language
}

func WarpError(err error, language, message string) IAPIError {
	return &APIError{err: err, language: language, message: message}
}

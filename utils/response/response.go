package response

import (
	"fmt"
	"strings"

	. "rela_recommend/i18n"
	"rela_recommend/utils/codec"
	"net/http"
	"rela_recommend/service"
	"rela_recommend/log"
)

type BaseResponse struct {
	Result    string `json:"result"`     //结果
	Errcode   string `json:"errcode"`    //错误码
	Errdesc   string `json:"errdesc"`    //错误描述/中文
	ErrdescEn string `json:"errdesc_en"` //错误描述/英文
}

type IBaseResponse interface {
	FormatErrcode(errcode, language, message string)
	FormatResule(result string)
}

func (this *BaseResponse) FormatErrcode(errcode, language, message string) {
	this.Errcode = errcode
	switch language {
	case codec.LANGUAGE_CHS:
		//this.Errdesc = format(Chs[errcode], message)
		this.Errdesc = Chs[errcode]
	case codec.LANGUAGE_CHT:
		this.Errdesc = Cht[errcode]
	case codec.LANGUAGE_CHT_TW:
		this.Errdesc = Cht_tw[errcode]
	case codec.LANGUAGE_EN:
		this.Errdesc = En[errcode]
	case codec.LANGUAGE_ES:
		this.Errdesc = Es[errcode]
	case codec.LANGUAGE_FR:
		this.Errdesc = Fr[errcode]
	case codec.LANGUAGE_TH:
		this.Errdesc = Th[errcode]
	case codec.LANGUAGE_JP:
		this.Errdesc = Jp[errcode]
	default:
		this.Errdesc = Chs[errcode]
	}
}

func (this *BaseResponse) FormatResule(result string) {
	this.Result = result
}

type BaseResponseV2 struct {
	Result    string `json:"result,omitempty"`     //结果
	Errcode   string `json:"errcode,omitempty"`    //错误码
	Errdesc   string `json:"errdesc,omitempty"`    //错误描述/中文
	ErrdescEn string `json:"errdesc_en,omitempty"` //错误描述/英文
}

type ResponseV2 struct {
	BaseResponseV2
	Data interface{} `json:"data,omitempty"` //数据
}

type IBaseResponseV2 interface {
	FormatErrcodeV2(errcode, language, message string)
	FormatResuleV2(result string)
}

func (this *BaseResponseV2) FormatErrcode(errcode, language, message string) {
	this.Errcode = errcode
	switch language {
	case codec.LANGUAGE_CHS:
		//this.Errdesc = format(Chs[errcode], message)
		this.Errdesc = Chs[errcode]
	case codec.LANGUAGE_CHT:
		this.Errdesc = Cht[errcode]
	case codec.LANGUAGE_CHT_TW:
		this.Errdesc = Cht_tw[errcode]
	case codec.LANGUAGE_EN:
		this.Errdesc = En[errcode]
	case codec.LANGUAGE_ES:
		this.Errdesc = Es[errcode]
	case codec.LANGUAGE_FR:
		this.Errdesc = Fr[errcode]
	case codec.LANGUAGE_TH:
		this.Errdesc = Th[errcode]
	case codec.LANGUAGE_JP:
		this.Errdesc = Jp[errcode]
	default:
		this.Errdesc = Chs[errcode]
	}
}
func (this *BaseResponseV2) FormatResule(result string) {
	this.Result = result
}

/*
alert = "%s访问了你%s"
message = 用户1#用户2
return 用户1访问了你用户2
*/
func format(alert, message string) string {
	length := strings.Count(alert, "%s")
	messages := strings.Split(message, "#")
	switch length {
	case 1:
		alert = fmt.Sprintf(alert, message)
	case 2:
		if len(messages) >= 2 {
			alert = fmt.Sprintf(alert, messages[0], messages[1])
		}
	case 3:
		if len(messages) >= 3 {
			alert = fmt.Sprintf(alert, messages[0], messages[1], messages[2])
		}
	default:
	}
	return alert
}


func FormatResponse(data interface{}, apiError service.IAPIError) (status int, res ResponseV2) {
	status = http.StatusBadRequest
	if apiError == nil {
		status = http.StatusOK
		res.Data = data
		return
	}
	err := apiError.GetError()
	message := apiError.GetMessage()
	language := apiError.GetLanguage()
	switch err {
	case nil:
		status = http.StatusOK
		res.Data = data
	case service.ErrInvaPara:
		status = http.StatusBadRequest
		res.FormatErrcode("param_error", language, message)
	case service.ErrInvaKey:
		res.FormatErrcode("require_login", language, message)
	default:
		status = http.StatusInternalServerError
		res.FormatErrcode("server_error", language, message)
	}

	if err != nil {
		log.Error(err.Error())
	}
	return
}

type ResponseV3 struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	TTL     int         `json:"ttl"`
}

func FormatResponseV3(data interface{}, err error) (status int, res ResponseV3) {
	if err == nil {
		status = http.StatusOK
		res.Data = data
		return
	} else {
		status = http.StatusBadRequest
		res.Code = 1
		res.Message = err.Error()
		log.Error(err.Error())
	}
	return
}

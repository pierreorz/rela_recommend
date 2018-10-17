package controllers

import (
	"net/http"
	"rela_recommend/service"
	"rela_recommend/response"
	"rela_recommend/log"
)

func formatResponse(data interface{}, apiError service.IAPIError) (status int, res response.ResponseV2) {
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
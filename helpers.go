package main

import (
	"github.com/ricardoecosta/weddingfeed/domain"
	"net/http"
)

func SendError(writer http.ResponseWriter, code int, message string) {
	errorMessage := &domain.ErrorMessage{Code: code, Message: message}
	errorMessageString := errorMessage.String()
	if len(errorMessageString) > 0 {
		http.Error(writer, errorMessageString, code)
	} else {
		http.Error(writer, "", http.StatusInternalServerError)
	}
}

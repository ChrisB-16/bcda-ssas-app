package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	ssas "github.com/CMSgov/bcda-ssas-app/ssas"
)

func WriteHttpError(w http.ResponseWriter, e ssas.ErrorResponse, errorStatus int) {
	fallbackMessage := fmt.Sprintf(`{"error": "%s", "error_description": "%s"}`, http.StatusText(http.StatusInternalServerError), http.StatusText(http.StatusInternalServerError))
	body, err := json.Marshal(e)

	if err != nil {
		http.Error(w, fallbackMessage, http.StatusInternalServerError)
	}

	w.WriteHeader(errorStatus)
	_, err = w.Write(body)

	if err != nil {
		http.Error(w, fallbackMessage, http.StatusInternalServerError)
	}
}

// Follow RFC 7591 format for input errors
func JsonError(w http.ResponseWriter, errorStatus int, statusText string, statusDescription string) {
	e := ssas.ErrorResponse{Error: statusText, ErrorDescription: statusDescription}

	WriteHttpError(w, e, errorStatus)

	ssas.Logger.Printf("%s; %s", statusDescription, statusText)
}

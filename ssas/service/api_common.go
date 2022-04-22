package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/CMSgov/bcda-ssas-app/ssas"
)

func WriteHttpError(w http.ResponseWriter, e ssas.ErrorResponse, fallbackMessage string, errorStatus int) {
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
	var statusDescriptionLength int = len(statusDescription)
	var hasStatusDescription bool = statusDescriptionLength > 0

	fallbackMessage := fmt.Sprintf(`{"error": "%s", "error_description": "%s"}`, http.StatusText(http.StatusInternalServerError), http.StatusText(http.StatusInternalServerError))

	if hasStatusDescription {
		e := ssas.ErrorResponse{Error: statusText, ErrorDescription: statusDescription}
		WriteHttpError(w, e, fallbackMessage, errorStatus)
		ssas.Logger.Printf("%s; %s", statusDescription, statusText)
	} else {
		var statusDescription string = http.StatusText(errorStatus)
		e := ssas.ErrorResponse{Error: statusText, ErrorDescription: statusDescription}
		WriteHttpError(w, e, fallbackMessage, errorStatus)
		ssas.Logger.Printf("%s; %s", statusDescription, statusText)
	}
}

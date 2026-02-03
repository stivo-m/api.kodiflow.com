package api

import (
	"encoding/json"
	"net/http"
)

type ApiResponse struct {
	Message string `json:"message,omitzero"`
	Data    any    `json:"data,omitzero"`
	Errors  any    `json:"errors,omitzero"`
	Code    int    `json:"-"`
}

func EncodeJSON[T any](w http.ResponseWriter, data T) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(data)
}

// Sends a successful http json response
func SendSuccessResponse(
	w http.ResponseWriter,
	code int,
	message string,
	data any,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	successResponse := ApiResponse{
		Message: message,
		Data:    data,
	}

	_ = json.NewEncoder(w).Encode(successResponse)
}

// Sends an error http json response
func SendErrorResponse(w http.ResponseWriter, code int, message string, details any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	errorResponse := ApiResponse{
		Code:    code,
		Message: message,
		Errors:  details,
	}

	_ = json.NewEncoder(w).Encode(errorResponse)
}

func ProcessUsecaseResponse(result *ApiResponse, w http.ResponseWriter) {
	if result.Code > 399 {
		SendErrorResponse(
			w,
			result.Code,
			result.Message,
			result.Errors,
		)
		return
	}

	SendSuccessResponse(w, result.Code, result.Message, result.Data)
}

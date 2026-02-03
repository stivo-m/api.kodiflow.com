package handlers

import (
	"net/http"

	"github.com/stivo-m/api.kodiflow.com/pkg/api"
)

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	api.SendSuccessResponse(
		w,
		http.StatusOK,
		"All services are running",
		map[string]any{"status": "ok"},
	)
}

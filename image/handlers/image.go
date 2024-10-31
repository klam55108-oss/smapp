package handlers

import (
	"context"
	"errors"
	"log"
	"net/http"

	commonhttp "smapp/common/http"
	"smapp/image/service"
)

func GenerateUploadForm(svc *service.GenerateUploadForm, imgSizeLimit int64) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Headers set by the gateway
		userID := r.Header.Get("X-User-Id")
		if userID == "" {
			commonhttp.JSONError(w, "Missing X-User-Id header", http.StatusUnauthorized)
			return
		}

		form, err := svc.GetForm(r.Context(), userID, imgSizeLimit)
		if errors.Is(err, context.DeadlineExceeded) {
			log.Println(err)
			commonhttp.JSONErrorWithDefaultMessage(w, http.StatusGatewayTimeout)
			return
		}
		if errors.Is(err, context.Canceled) {
			// client disconnected
			log.Println(err)
			return
		}
		if err != nil {
			commonhttp.JSONError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		commonhttp.JSONResponse(w, form, http.StatusOK)
	})
}

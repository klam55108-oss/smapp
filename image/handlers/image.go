package handlers

import (
	"context"
	"errors"
	"log"
	"net/http"

	"smapp/common/jsonresp"
	"smapp/image/service"
)

func GenerateUploadForm(svc *service.GenerateUploadForm, imgPurpose string, imgSizeLimit int64) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Headers set by the gateway
		userID := r.Header.Get("X-User-Id")
		if userID == "" {
			jsonresp.Error(w, "Missing X-User-Id header", http.StatusUnauthorized)
			return
		}

		form, err := svc.GetForm(r.Context(), imgPurpose, userID, imgSizeLimit)
		if errors.Is(err, context.DeadlineExceeded) {
			log.Println(err)
			jsonresp.ErrorWithDefaultMessage(w, http.StatusGatewayTimeout)
			return
		}
		if errors.Is(err, context.Canceled) {
			// client disconnected
			log.Println(err)
			return
		}
		if err != nil {
			jsonresp.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"status": "success",
			"data":   form,
		}
		jsonresp.Response(w, response, http.StatusOK)
	})
}

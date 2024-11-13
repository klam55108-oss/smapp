package handlers

import (
	"context"
	"errors"
	"log"
	"net/http"
	commonmw "smapp/common/middleware"

	"smapp/common/jsonresp"
	"smapp/image/service"
)

func GenerateUploadForm(svc *service.GenerateUploadForm, imgPurpose string, imgSizeLimit int64) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := commonmw.GetUserID(r.Context())
		if err != nil {
			log.Println(err)
			jsonresp.ErrorWithDefaultMessage(w, http.StatusInternalServerError)
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

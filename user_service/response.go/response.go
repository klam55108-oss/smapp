package response

import (
	"encoding/json"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func JSONErrorFromMap(w http.ResponseWriter, jsonMap map[string]interface{}, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(jsonMap)
}

func JSONError(w http.ResponseWriter, error string, code int) {
	JSONErrorFromMap(w, map[string]interface{}{"status": "error", "message": error}, code)
}

func JSONErrorWithDefaultMessage(w http.ResponseWriter, code int) {
	JSONError(w, http.StatusText(code), code)
}

func JSONValidationError(w http.ResponseWriter, errors validation.Errors, code int) {
	result := map[string]interface{}{
		"status":  "error",
		"message": "Validation failed",
		"errors":  make(map[string]string),
	}
	for key, err := range errors {
		result["errors"].(map[string]string)[key] = err.Error()
	}
	JSONErrorFromMap(w, result, code)
}

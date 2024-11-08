package jsonresp

import (
	"encoding/json"
	"net/http"
)

func Response(w http.ResponseWriter, jsonMap map[string]interface{}, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(jsonMap)
}

func Error(w http.ResponseWriter, error string, code int) {
	Response(w, map[string]interface{}{"status": "error", "message": error}, code)
}

func ErrorWithDefaultMessage(w http.ResponseWriter, code int) {
	Error(w, http.StatusText(code), code)
}

func ValidationError(w http.ResponseWriter, errors map[string]error, code int) {
	result := map[string]interface{}{
		"status":  "error",
		"message": "Validation failed",
		"errors":  make(map[string]string),
	}
	for key, err := range errors {
		result["errors"].(map[string]string)[key] = err.Error()
	}
	Response(w, result, code)
}

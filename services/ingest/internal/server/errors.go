package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
)

type errorResponse struct {
	Error  string `json:"error"`
	Detail string `json:"detail,omitempty"`
	Dep    string `json:"dep,omitempty"`
}

func writeError(w http.ResponseWriter, status int, msg string, detail any) {
	body := errorResponse{Error: msg}
	if detail != nil && status < 500 {
		switch v := detail.(type) {
		case error:
			body.Detail = v.Error()
		case int64:
			body.Detail = strconv.FormatInt(v, 10)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)

	if status >= 500 {
		slog.Error("server error", "status", status, "msg", msg, "detail", detail)
	}
}

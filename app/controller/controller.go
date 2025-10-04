package controller

import (
	"encoding/json"
	"net/http"
	"synk/gateway/app/util"
)

type Response struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
	Info  any    `json:"info"`
	List  []any  `json:"list"`
}

type ResponseHeader struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
}

func EnableCors(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

func WriteErrorResponse(w http.ResponseWriter, response any, route string, message string, status int) {
	util.LogRoute(route, message)

	jsonResp, _ := json.Marshal(response)

	w.WriteHeader(status)
	w.Write(jsonResp)
}

func WriteSuccessResponse(w http.ResponseWriter, response any) {
	jsonResp, _ := json.Marshal(response)

	w.WriteHeader(http.StatusOK)
	w.Write(jsonResp)
}

func SetJsonContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}

package controller

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"synk/gateway/app/util"
	"time"
)

type Response struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
	Info  any    `json:"info"`
	List  []any  `json:"list"`
}

type HandleAuthResponse struct {
	Resource ResponseHeader `json:"resource"`
}

type ResponseHeader struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
}

const AUTH_TIMEOUT = time.Second * 5

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

func Cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		response := HandleAuthResponse{
			Resource: ResponseHeader{
				Ok: true,
			},
		}

		authEndpoint := os.Getenv("AUTH_ENDPOINT")

		if authEndpoint == "" {
			response.Resource.Ok = false
			response.Resource.Error = "auth server is not configured"

			WriteErrorResponse(w, response, "/", response.Resource.Error, http.StatusInternalServerError)

			return
		}

		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			response.Resource.Ok = false
			response.Resource.Error = "Bearer Authorization header is required"

			WriteErrorResponse(w, response, "/", response.Resource.Error, http.StatusBadRequest)

			return
		}

		authReq, authReqErr := http.NewRequest("GET", strings.TrimSuffix(authEndpoint, "/")+"/users/check", nil)
		if authReqErr != nil {
			response.Resource.Ok = false
			response.Resource.Error = "error while setting auth server"

			WriteErrorResponse(w, response, "/", response.Resource.Error, http.StatusInternalServerError)

			return
		}

		authReq.Header.Set("Authorization", authHeader)
		authReq.Header.Set("Accept", "application/json")

		client := &http.Client{
			Timeout: AUTH_TIMEOUT,
		}
		authResp, authRespErr := client.Do(authReq)
		if authRespErr != nil {
			response.Resource.Ok = false
			response.Resource.Error = "error while contacting auth server"

			WriteErrorResponse(w, response, "/", response.Resource.Error, http.StatusInternalServerError)

			return
		}
		defer authResp.Body.Close()

		if authResp.StatusCode != http.StatusOK {
			var authCheckResponse HandleAuthResponse

			bodyBytes, readErr := io.ReadAll(authResp.Body)

			if readErr != nil {
				response.Resource.Ok = false
				response.Resource.Error = "error while parsing auth server response"

				WriteErrorResponse(w, response, "/", response.Resource.Error, http.StatusInternalServerError)

				return
			}

			if err := json.Unmarshal(bodyBytes, &authCheckResponse); err != nil {
				response.Resource.Ok = false
				response.Resource.Error = "error while decoding auth server response"

				WriteErrorResponse(w, response, "/", response.Resource.Error, http.StatusInternalServerError)

				return
			}

			if !authCheckResponse.Resource.Ok {
				response.Resource.Ok = false
				response.Resource.Error = authCheckResponse.Resource.Error

				WriteErrorResponse(w, response, "/", response.Resource.Error, authResp.StatusCode)

				return
			}

			return
		}

		next.ServeHTTP(w, r)
	})
}

package controller

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"synk/gateway/app/util"
	"time"

	"github.com/getsentry/sentry-go"
)

type Response struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
	Info  any    `json:"info"`
	List  []any  `json:"list"`
}

type UserAuthDataResponse struct {
	UserId int `json:"user_id"`
}

type HandleAuthResponse struct {
	Resource ResponseHeader       `json:"resource"`
	Data     UserAuthDataResponse `json:"user"`
}

type ResponseHeader struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
}

type ContextKey string

const AUTH_TIMEOUT = time.Second * 5
const SENTRY_LOG_TIMEOUT = time.Second * 5
const CONTEXT_USER_ID_KEY ContextKey = "user_id"

func WriteErrorResponse(w http.ResponseWriter, response any, route string, message string, status int) {
	defer sentry.Flush(SENTRY_LOG_TIMEOUT)

	util.LogRoute(route, message)

	sentry.CaptureMessage("error(@gateway" + route + "): " + message)

	jsonResp, _ := json.Marshal(response)

	w.WriteHeader(status)
	w.Write(jsonResp)
}

func WriteSuccessResponse(w http.ResponseWriter, response any) {
	defer sentry.Flush(SENTRY_LOG_TIMEOUT)

	jsonResp, _ := json.Marshal(response)

	sentry.CaptureMessage("success(@gateway): " + string(jsonResp))

	w.WriteHeader(http.StatusOK)
	w.Write(jsonResp)
}

func SetJsonContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}

func NewServiceClient() *http.Client {
	env := os.Getenv("ENV")

	if env == "production" {
		return &http.Client{
			Timeout: AUTH_TIMEOUT,
		}
	}

	caCert, caErr := os.ReadFile("/cert/rootCA.pem")
	if caErr != nil {
		util.LogRoute("/", "error reading root CA file: "+caErr.Error())
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		RootCAs: caCertPool,
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   AUTH_TIMEOUT,
	}

	return client
}

var authClient = NewServiceClient()

func Cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var allowedOriginsMap = map[string]struct{}{
			strings.TrimSuffix(os.Getenv("WEB_ENDPOINT"), "/"): {},
		}

		w.Header().Set("Access-Control-Allow-Credentials", "true")

		origin := r.Header.Get("Origin")
		if _, ok := allowedOriginsMap[origin]; ok {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
		}

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

		authResp, authRespErr := authClient.Do(authReq)
		if authRespErr != nil {
			response.Resource.Ok = false
			response.Resource.Error = "error while contacting auth server"

			WriteErrorResponse(w, response, "/", response.Resource.Error, http.StatusInternalServerError)

			return
		}
		defer authResp.Body.Close()

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

		if authResp.StatusCode != http.StatusOK && !authCheckResponse.Resource.Ok {
			response.Resource.Ok = false
			response.Resource.Error = authCheckResponse.Resource.Error

			WriteErrorResponse(w, response, "/", response.Resource.Error, authResp.StatusCode)

			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, CONTEXT_USER_ID_KEY, authCheckResponse.Data.UserId)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

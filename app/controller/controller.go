package controller

type Response struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
	Info  any    `json:"info"`
	List  []any  `json:"list"`
}

package utils

type APIResponse struct {
	Status  string      `json:"string"`
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Errors  interface{} `json:"errors,omitempty"`
}

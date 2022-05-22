package response

import (
	"encoding/json"
	"net/http"
)

type ApiResponse struct {
	ResponseWriter http.ResponseWriter
	Target         string // service name
}

// swagger:model EntityCreated
type EntityCreated struct {
	ID string `json:"id"`
}

func NewResponseJSON(w http.ResponseWriter, target string) *ApiResponse {
	w.Header().Add("Content-Type", "application/json")
	return &ApiResponse{
		ResponseWriter: w,
		Target:         target,
	}
}

func NewResponse(w http.ResponseWriter, target string, headers map[string]string) *ApiResponse {
	for key, value := range headers {
		w.Header().Add(key, value)
	}
	return &ApiResponse{
		ResponseWriter: w,
		Target:         target,
	}
}

func (responser *ApiResponse) AddHeader(key string, value string) *ApiResponse {
	responser.ResponseWriter.Header().Add(key, value)
	return responser
}

func (responser *ApiResponse) DeleteHeader(key string) *ApiResponse {
	responser.ResponseWriter.Header().Del(key)
	return responser
}

func (responser *ApiResponse) writeResponse(status int, body interface{}) {
	if body == nil && !isSuccessCode(status) {
		body = MakeApiError(getCodeText(status), "unknown error", responser.Target)
	}

	respBody, err := json.Marshal(body)
	if err != nil {
		responser.writeResponse(http.StatusInternalServerError, MakeApiError(getCodeText(status), "unknown error", responser.Target))
		return
	}

	responser.ResponseWriter.WriteHeader(status)
	if body != nil {
		responser.ResponseWriter.Write(respBody)
	}
}

func (responser *ApiResponse) Ok(entity interface{}) {
	responser.writeResponse(http.StatusOK, entity)
}

func (responser *ApiResponse) Created(entity interface{}) {
	responser.writeResponse(http.StatusCreated, entity)
}

func (responser *ApiResponse) Accepted(entity interface{}) {
	responser.writeResponse(http.StatusAccepted, entity)
}

func (responser *ApiResponse) NoContent() {
	responser.writeResponse(http.StatusNoContent, nil)
}

func (responser *ApiResponse) Unauthorized(entity interface{}) {
	responser.writeResponse(http.StatusUnauthorized, entity)
}

func (responser *ApiResponse) BadRequest(entity interface{}) {
	responser.writeResponse(http.StatusBadRequest, entity)
}

func (responser *ApiResponse) Forbidden(entity interface{}) {
	responser.writeResponse(http.StatusForbidden, entity)
}

func (responser *ApiResponse) NotFound(entity interface{}) {
	responser.writeResponse(http.StatusNotFound, entity)
}

func (responser *ApiResponse) MethodNotAllowed(entity interface{}) {
	responser.writeResponse(http.StatusMethodNotAllowed, entity)
}

func (responser *ApiResponse) Conflict(entity interface{}) {
	responser.writeResponse(http.StatusConflict, entity)
}

func (responser *ApiResponse) InternalServerError(entity interface{}) {
	responser.writeResponse(http.StatusInternalServerError, entity)
}

func (responser *ApiResponse) NotImplemented(entity interface{}) {
	responser.writeResponse(http.StatusNotImplemented, entity)
}

func (responser *ApiResponse) ServiceUnavailable(entity interface{}) {
	responser.writeResponse(http.StatusServiceUnavailable, entity)
}

func isSuccessCode(status int) bool {
	return status/100 == 2
}

func getCodeText(status int) string {
	return http.StatusText(status)
}

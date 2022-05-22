package response

// swagger:model ApiError
type ApiError struct {
	// Code operation
	// in: string
	Code string `json:"code"`
	// Error description
	Message string `json:"message,omitempty"`
	// API Target
	Target string `json:"target,omitempty"`
	// Anything data
	Context interface{} `json:"context,omitempty"`
	// TraceID
	TraceId string `json:"traceId,omitempty"`
	// Additional Errors
	Errors []ApiError `json:"errors,omitempty"`
}

func (e ApiError) Error() string {
	result := "Err: " + string(e.Code) + " " + e.Message
	if e.Target != "" {
		result += "Service: " + e.Target
	}
	if e.TraceId != "" {
		result += "TraceID: " + e.TraceId
	}
	return result
}

func MakeApiError(code string, message string, target string) ApiError {
	return ApiError{
		Code:    code,
		Message: message,
		Target:  target,
	}
}

func MakeSimpleApiError(code string, message string) ApiError {
	return ApiError{
		Code:    code,
		Message: message,
	}
}

func (err *ApiError) SetTarget(target string) {
	if target != "" {
		err.Target = target
	}
}

func (err *ApiError) SetCode(code string) {
	if code != "" {
		err.Code = code
	}
}

func (err *ApiError) SetTraceID(traceId string) {
	if traceId != "" {
		err.TraceId = traceId
	}
}

func (err *ApiError) AddContextError(newSubErr ApiError) {
	err.Errors = append(err.Errors, newSubErr)
}

func (err *ApiError) SetContext(newContext string) {
	err.Context = newContext
}

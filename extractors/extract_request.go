package extractors

import (
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
)

const (
	paramExtracting = "ex_param"
)

type extractor func(string, *http.Request) interface{}

func RequestToType(request *http.Request, data interface{}, parseQuery, parseParams bool) (interface{}, error) {
	if err := json.NewDecoder(request.Body).Decode(&data); err != nil {
		return nil, errors.New("can not unmarshalled request")
	}

	var extractorFunc extractor

	if parseQuery {
		extractorFunc = extractFromQuery
	}

	if parseParams {
		extractorFunc = extractFromPathVariables
	}

	value := reflect.Indirect(reflect.ValueOf(data))

	if err := prepareInlineStructFields(request, value, extractorFunc); err != nil {
		return nil, err
	}

	return data, nil
}

func extractFromQuery(paramName string, request *http.Request) interface{} {
	return request.URL.Query().Get(paramName)
}

func extractFromPathVariables(paramName string, request *http.Request) interface{} {
	vars := request.Context().Value(0).(map[string]interface{})
	return vars[paramName].(string)
}

func extractFromHeaders(paramName string, request *http.Request) interface{} {
	return request.Header.Get(paramName)
}

func prepareInlineStructFields(request *http.Request, value reflect.Value, preparator extractor) error {
	for i := 0; i < value.NumField(); i++ {
		val := reflect.Indirect(value.Field(i).Addr())
		if val.Kind() == reflect.Struct {
			prepareInlineStructFields(request, val, preparator)
		} else {
			parsedTag := value.Type().Field(i).Tag.Get(paramExtracting)
			if parsedTag != "" {
				dataQuery := preparator(parsedTag, request)
				if dataQuery != "" {
					value.Set(reflect.ValueOf(dataQuery)) // TODO: need format to concrete type
				} else {
					return errors.New("in request does not exist query param with name: " + parsedTag)
				}
			}
		}
	}
	return nil
}

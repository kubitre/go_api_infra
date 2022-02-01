package extractors

import (
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
)

const (
	tagParsing = "query"
)

// TODO: make abstract type interface{}
type extractor func(string, *http.Request) string

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

	value := reflect.ValueOf(data)

	if err := prepareInlineStructFields(request, value, extractorFunc); err != nil {
		return nil, err
	}

	return data, nil
}

func extractFromQuery(paramName string, request *http.Request) string {
	return request.URL.Query().Get(paramName)
}

func extractFromPathVariables(paramName string, request *http.Request) string {
	vars := request.Context().Value(0).(map[string]interface{})
	return vars[paramName].(string)
}

func prepareInlineStructFields(request *http.Request, value reflect.Value, preparator extractor) error {
	for i := 0; i < value.NumField(); i++ {
		val := value.Field(i).Addr()
		if val.Kind() == reflect.Struct {
			prepareInlineStructFields(request, val, preparator)
		} else {
			parsedTag := val.Type().Field(i).Tag.Get(tagParsing)
			if parsedTag != "" {
				dataQuery := request.URL.Query().Get(parsedTag)
				if dataQuery != "" {
					value.SetString(dataQuery)
				} else {
					return errors.New("in request does not exist query param with name: " + parsedTag)
				}
			}
		}
	}
	return nil
}

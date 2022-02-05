package extractors

import (
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
)

type ParserType string

const (
	ParseFromHeaders   ParserType = "parse_headers"
	ParseFromQuery     ParserType = "parse_query"
	ParsePathVariables ParserType = "parse_path_variables"
	paramExtracting    string     = "ex_param"
)

type extractor func(string, *http.Request) interface{}

func RequestToType(request *http.Request, data interface{}, tartgetParamsPlaces ...ParserType) (interface{}, error) {
	if err := json.NewDecoder(request.Body).Decode(&data); err != nil {
		return nil, errors.New("can not unmarshalled request")
	}

	var extractorFuncs []extractor

	for _, extractorParam := range tartgetParamsPlaces {
		switch extractorParam {
		case ParseFromHeaders:
			extractorFuncs = append(extractorFuncs, extractFromHeaders)
		case ParseFromQuery:
			extractorFuncs = append(extractorFuncs, extractFromQuery)
		case ParsePathVariables:
			extractorFuncs = append(extractorFuncs, extractFromPathVariables)
		}
	}

	value := reflect.Indirect(reflect.ValueOf(data))

	if err := prepareInlineStructFields(request, value, extractorFuncs); err != nil {
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

func prepareInlineStructFields(request *http.Request, value reflect.Value, preparators []extractor) error {
	for i := 0; i < value.NumField(); i++ {
		val := reflect.Indirect(value.Field(i).Addr())
		if val.Kind() == reflect.Struct {
			prepareInlineStructFields(request, val, preparators)
		} else {
			parsedTag := value.Type().Field(i).Tag.Get(paramExtracting)
			if parsedTag != "" {
				parsedValue := getValueInAllExtractors(parsedTag, request, preparators)
				if parsedValue.Kind() == reflect.ValueOf(nil).Kind() {
					return errors.New("in request does not exist query param with name: " + parsedTag)
				}
				value.Set(parsedValue)
			}
		}
	}
	return nil
}

func getValueInAllExtractors(parsedTag string, request *http.Request, extractors []extractor) reflect.Value {
	for _, extractFunc := range extractors {
		if res := extractFunc(parsedTag, request); res != nil {
			return reflect.ValueOf(res)
		}
	}
	return reflect.ValueOf(nil)
}

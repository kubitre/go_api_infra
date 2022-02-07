package extractors

import (
	"encoding/json"
	"errors"
	"net/http"
	"reflect"

	"github.com/spf13/cast"
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
				setValueToType(val, parsedValue.String())
			}
		}
	}
	return nil
}

func setValueToType(val reflect.Value, parsedValue string) error {
	if val.CanSet() {
		return errors.New("can not setting value to this field: " + val.Addr().String())
	}
	switch val.Kind() {
	case reflect.Int:
		val.Set(reflect.ValueOf(cast.ToInt(parsedValue)))
	case reflect.Int16:
		val.Set(reflect.ValueOf(cast.ToInt16(parsedValue)))
	case reflect.Int32:
		val.Set(reflect.ValueOf(cast.ToInt32(parsedValue)))
	case reflect.Int64:
		val.Set(reflect.ValueOf(cast.ToInt64(parsedValue)))
	case reflect.Float32:
		val.Set(reflect.ValueOf(cast.ToFloat32(parsedValue)))
	case reflect.Float64:
		val.Set(reflect.ValueOf(cast.ToFloat64(parsedValue)))
	case reflect.String:
		val.Set(reflect.ValueOf(parsedValue))
	case reflect.Bool:
		val.Set(reflect.ValueOf(cast.ToBool(parsedValue)))
	default:
		return errors.New("error while setting up")
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

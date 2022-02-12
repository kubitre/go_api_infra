package extractors

import (
	"encoding/json"
	"errors"
	"net/http"
	"reflect"

	"github.com/gorilla/mux"
	"github.com/kubitre/go_api_infra/response"
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

/**
RequestToType - extract from http.Request payload data from:
1. Body (prefer)
2. PathVariables(optional)
3. QueryParams(optional)
4. Headers(optional)
*/
func RequestToType(request *http.Request, writer http.ResponseWriter, target string, data interface{}, tartgetParamsPlaces ...ParserType) (interface{}, error) {
	if err := json.NewDecoder(request.Body).Decode(&data); err != nil {
		if len(tartgetParamsPlaces) == 0 {
			response.NewResponseJSON(writer, target).BadRequest(response.ApiError{Code: "WEB_INPUT_ERROR", Message: "can not deserialize input request", Target: target, Context: "Read spec at: /swagger"})
			return nil, err
		}
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
		response.NewResponseJSON(writer, target).BadRequest(
			response.ApiError{Code: "WEB_INPUT_ERROR", Message: "can not deserialize input request", Target: target, Context: "Read spec at: /swagger"},
		)
		return nil, err
	}

	return data, nil
}

func extractFromQuery(paramName string, request *http.Request) interface{} {
	return request.URL.Query().Get(paramName)
}

func extractFromPathVariables(paramName string, request *http.Request) interface{} {
	vars := mux.Vars(request)
	return vars[paramName]
}

func extractFromHeaders(paramName string, request *http.Request) interface{} {
	return request.Header.Get(paramName)
}

func prepareInlineStructFields(request *http.Request, value reflect.Value, preparators []extractor) error {
	for i := 0; i < value.NumField(); i++ {
		val := reflect.Indirect(value.Field(i))
		valAddress := value.Field(i).Addr()
		if val.Kind() == reflect.Struct {
			prepareInlineStructFields(request, val, preparators)
		} else {
			parsedTag := value.Type().Field(i).Tag.Get(paramExtracting)
			if parsedTag != "" {
				parsedValue := getValueInAllExtractors(parsedTag, request, preparators)
				if parsedValue.Kind() == reflect.ValueOf(nil).Kind() {
					return errors.New("in request does not exist query param with name: " + parsedTag)
				}
				if err := setValueToType(val, valAddress, parsedValue.String()); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func setValueToType(val reflect.Value, valToSet reflect.Value, parsedValue string) error {
	indirect := reflect.Indirect(valToSet)
	if valToSet.CanSet() {
		return errors.New("can not setting value to this field: " + val.Addr().String())
	}
	switch val.Kind() {
	case reflect.Int:
		indirect.Set(reflect.ValueOf(cast.ToInt(parsedValue)))
	case reflect.Int16:
		indirect.Set(reflect.ValueOf(cast.ToInt16(parsedValue)))
	case reflect.Int32:
		indirect.Set(reflect.ValueOf(cast.ToInt32(parsedValue)))
	case reflect.Int64:
		indirect.Set(reflect.ValueOf(cast.ToInt64(parsedValue)))
	case reflect.Float32:
		indirect.Set(reflect.ValueOf(cast.ToFloat32(parsedValue)))
	case reflect.Float64:
		indirect.Set(reflect.ValueOf(cast.ToFloat64(parsedValue)))
	case reflect.String:
		indirect.Set(reflect.ValueOf(parsedValue))
	case reflect.Bool:
		indirect.Set(reflect.ValueOf(cast.ToBool(parsedValue)))
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

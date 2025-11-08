package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Lexographics/autorpc"
	"github.com/Lexographics/autorpc/types"
)

func Slice(ctx context.Context, params []string) ([]string, error) {
	res := []string{}
	for _, param := range params {
		res = append(res, "Hello, "+param+"!")
	}
	return res, nil
}

type MapParams struct {
	Map map[string]string `json:"map" validate:"required"`
}

func MapStruct(ctx context.Context, params MapParams) (MapParams, error) {
	res := map[string]string{}
	for key, value := range params.Map {
		res[key] = "Hello, " + value + "!"
	}
	return MapParams{Map: res}, nil
}

func Map(ctx context.Context, params map[string]string) (map[string]string, error) {
	res := map[string]string{}
	for key, value := range params {
		res[key] = "Hello, " + value + "!"
	}
	return res, nil
}

type TimeParams struct {
	Time *types.Time `json:"time" validate:"required"`
}

func TimeStruct(ctx context.Context, params TimeParams) (types.Time, error) {
	return types.Time{Time: params.Time.Time.Add(time.Hour)}, nil
}

func Time(ctx context.Context, params types.Time) (types.Time, error) {
	return types.Time{Time: params.Time.Add(time.Hour)}, nil
}

type DurationParams struct {
	Duration types.Duration `json:"duration" validate:"required"`
}

func DurationStruct(ctx context.Context, params DurationParams) (types.Duration, error) {
	return types.Duration{Duration: params.Duration.Duration + time.Hour}, nil
}

func Duration(ctx context.Context, params types.Duration) (types.Duration, error) {
	return types.Duration{Duration: params.Duration + time.Hour}, nil
}

type NestedStructParams struct {
	Name   string              `json:"name" validate:"required"`
	Nested *NestedStructParams `json:"nested"`
}

func NestedStruct(ctx context.Context, params NestedStructParams) (NestedStructParams, error) {
	return params, nil
}

type DoublePointerParams struct {
	Value **string `json:"value"`
}

func DoublePointerStruct(ctx context.Context, params DoublePointerParams) (DoublePointerParams, error) {
	return params, nil
}

func DoublePointer(ctx context.Context, params **string) (**string, error) {
	return params, nil
}

func DoubleSlice(ctx context.Context, params [][]string) ([][]string, error) {
	return params, nil
}

func AnonymousStruct(ctx context.Context, params struct {
	Name string `json:"name" validate:"required"`
}) (result struct {
	Result string `json:"result"`
}, err error) {
	result.Result = "Hello, " + params.Name + "!"
	return result, nil
}

type NumberEnum int

const (
	NumberEnum0 NumberEnum = iota
	NumberEnum1
	NumberEnum2
	NumberEnum3
)

type EnumParams struct {
	Value NumberEnum `json:"value"`
}

func EnumStruct(ctx context.Context, params EnumParams) (NumberEnum, error) {
	return params.Value, nil
}

func Enum(ctx context.Context, params NumberEnum) (NumberEnum, error) {
	return params, nil
}

func (NumberEnum) UnmarshalKind() string {
	return "string"
}

func (n *NumberEnum) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case nil:
		return nil
	case float64:
		*n = NumberEnum(value)
	case string:
		switch value {
		case "zero", "":
			*n = NumberEnum0
		case "one":
			*n = NumberEnum1
		case "two":
			*n = NumberEnum2
		case "three":
			*n = NumberEnum3
		}
		return nil
	}
	return nil
}

func main() {
	server := autorpc.NewServer()

	testsGroup := server.Group("tests.")
	autorpc.RegisterMethod(testsGroup, "slice", Slice)
	autorpc.RegisterMethod(testsGroup, "map", Map)
	autorpc.RegisterMethod(testsGroup, "map-struct", MapStruct)
	autorpc.RegisterMethod(testsGroup, "time", Time)
	autorpc.RegisterMethod(testsGroup, "time-struct", TimeStruct)
	autorpc.RegisterMethod(testsGroup, "duration", Duration)
	autorpc.RegisterMethod(testsGroup, "duration-struct", DurationStruct)
	autorpc.RegisterMethod(testsGroup, "nested-struct", NestedStruct)
	autorpc.RegisterMethod(testsGroup, "double-pointer", DoublePointer)
	autorpc.RegisterMethod(testsGroup, "double-pointer-with-struct-params", DoublePointerStruct)
	autorpc.RegisterMethod(testsGroup, "double-slice", DoubleSlice)
	autorpc.RegisterMethod(testsGroup, "anonymous-struct", AnonymousStruct)
	autorpc.RegisterMethod(testsGroup, "enum", Enum)
	autorpc.RegisterMethod(testsGroup, "enum-struct", EnumStruct)

	http.Handle("/rpc", autorpc.HTTPHandler(server))
	http.Handle("/spec", autorpc.SpecUIHandler(server))
	http.Handle("/spec.json", autorpc.SpecJSONHandler(server))

	log.Println("Server started on port 8080")
	http.ListenAndServe(":8080", nil)
}

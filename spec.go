package autorpc

import (
	"fmt"
	"reflect"
	"strings"
)

type FieldInfo struct {
	Name            string      `json:"name"`
	JSONName        string      `json:"jsonName,omitempty"`
	Type            string      `json:"type"`
	Kind            string      `json:"kind"`
	Required        bool        `json:"required,omitempty"`
	ValidationRules []string    `json:"validationRules,omitempty"`
	IsArray         bool        `json:"isArray,omitempty"`
	IsPointer       bool        `json:"isPointer,omitempty"`
	ElementType     string      `json:"elementType,omitempty"` // for arrays/slices/pointers
	Fields          []FieldInfo `json:"fields,omitempty"`      // nested fields if this is a struct type
}

type TypeInfo struct {
	Name        string      `json:"name"`
	Kind        string      `json:"kind"`
	IsArray     bool        `json:"isArray,omitempty"`
	IsPointer   bool        `json:"isPointer,omitempty"`
	ElementType string      `json:"elementType,omitempty"` // for arrays/slices/pointers
	Fields      []FieldInfo `json:"fields,omitempty"`      // nested fields if this is a struct type
}

type MethodInfo struct {
	Name   string   `json:"name"`
	Params TypeInfo `json:"params"`
	Result TypeInfo `json:"result"`
}

// GetMethodSpecs returns information about all registered RPC methods.
// This can be used to generate API specifications
//
// Example usage:
//
//	specs := server.GetMethodSpecs()
//	for _, spec := range specs {
//	    fmt.Printf("Method: %s\n", spec.Name)
//	    fmt.Printf("Params: %+v\n", spec.Params)
//	    fmt.Printf("Result: %+v\n", spec.Result)
//	}
func (s *Server) GetMethodSpecs() []MethodInfo {
	var specs []MethodInfo

	s.methods.Range(func(key, value interface{}) bool {
		methodName := key.(string)
		handler := value.(methodHandler)

		fnType := handler.fnValue.Type()

		paramType := fnType.In(0)
		resultType := fnType.Out(0)

		spec := MethodInfo{
			Name:   methodName,
			Params: extractTypeInfo(paramType),
			Result: extractTypeInfo(resultType),
		}

		specs = append(specs, spec)
		return true
	})

	return specs
}

func extractTypeInfo(typ reflect.Type) TypeInfo {
	info := TypeInfo{}

	for typ.Kind() == reflect.Ptr {
		info.IsPointer = true
		typ = typ.Elem()
	}

	for typ.Kind() == reflect.Array || typ.Kind() == reflect.Slice {
		info.IsArray = true
		elemType := typ.Elem()
		for elemType.Kind() == reflect.Ptr {
			elemType = elemType.Elem()
		}
		info.ElementType = getTypeName(elemType)
		typ = elemType
	}

	info.Name = getTypeName(typ)
	info.Kind = typ.Kind().String()

	if typ.Kind() == reflect.Struct {
		info.Fields = extractFields(typ)
	}

	return info
}

func extractFields(typ reflect.Type) []FieldInfo {
	var fields []FieldInfo

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		if !field.IsExported() {
			continue
		}

		fieldInfo := FieldInfo{
			Name: field.Name,
		}

		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}
		if jsonTag != "" {
			parts := strings.Split(jsonTag, ",")
			fieldInfo.JSONName = parts[0]
		} else {
			fieldInfo.JSONName = field.Name
		}

		validateTag := field.Tag.Get("validate")
		if validateTag != "" {
			fieldInfo.ValidationRules = strings.Split(validateTag, ",")
			for _, rule := range fieldInfo.ValidationRules {
				if strings.HasPrefix(strings.TrimSpace(rule), "required") {
					fieldInfo.Required = true
					break
				}
			}
		}

		fieldType := field.Type

		for fieldType.Kind() == reflect.Ptr {
			fieldInfo.IsPointer = true
			fieldType = fieldType.Elem()
		}

		for fieldType.Kind() == reflect.Array || fieldType.Kind() == reflect.Slice {
			fieldInfo.IsArray = true
			elemType := fieldType.Elem()
			for elemType.Kind() == reflect.Ptr {
				elemType = elemType.Elem()
			}
			fieldInfo.ElementType = getTypeName(elemType)
			fieldType = elemType
		}

		fieldInfo.Type = getTypeName(fieldType)
		fieldInfo.Kind = fieldType.Kind().String()

		if fieldType.Kind() == reflect.Struct {
			fieldInfo.Fields = extractFields(fieldType)
		}

		fields = append(fields, fieldInfo)
	}

	return fields
}

func getTypeName(typ reflect.Type) string {
	switch typ.Kind() {
	case reflect.Bool:
		return "bool"
	case reflect.Int:
		return "int"
	case reflect.Int8:
		return "int8"
	case reflect.Int16:
		return "int16"
	case reflect.Int32:
		return "int32"
	case reflect.Int64:
		return "int64"
	case reflect.Uint:
		return "uint"
	case reflect.Uint8:
		return "uint8"
	case reflect.Uint16:
		return "uint16"
	case reflect.Uint32:
		return "uint32"
	case reflect.Uint64:
		return "uint64"
	case reflect.Float32:
		return "float32"
	case reflect.Float64:
		return "float64"
	case reflect.String:
		return "string"
	case reflect.Struct:
		if typ.Name() != "" {
			return typ.Name()
		}
		return "struct"
	case reflect.Array, reflect.Slice:
		return fmt.Sprintf("[]%s", getTypeName(typ.Elem()))
	case reflect.Map:
		return fmt.Sprintf("map[%s]%s", getTypeName(typ.Key()), getTypeName(typ.Elem()))
	default:
		if typ.Name() != "" {
			return typ.Name()
		}
		return typ.String()
	}
}

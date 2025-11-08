package autorpc

import (
	"crypto/sha256"
	"encoding/hex"
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
	ArrayDepth      int         `json:"arrayDepth,omitempty"` // 0 = not an array, 1 = []T, 2 = [][]T, etc.
	IsPointer       bool        `json:"isPointer,omitempty"`
	PointerDepth    int         `json:"pointerDepth,omitempty"` // 0 = not a pointer, 1 = *T, 2 = **T, etc.
	ElementType     string      `json:"elementType,omitempty"`  // string representation of element type (hint)
	KeyType         string      `json:"keyType,omitempty"`      // key type for map types
	ValueType       string      `json:"valueType,omitempty"`    // value type for map types
	Fields          []FieldInfo `json:"fields,omitempty"`       // nested fields if this is a struct type
}

type TypeInfo struct {
	Name         string      `json:"name"`              // struct name only (without package)
	Package      string      `json:"package,omitempty"` // package path
	Kind         string      `json:"kind"`
	IsArray      bool        `json:"isArray,omitempty"`
	ArrayDepth   int         `json:"arrayDepth,omitempty"` // 0 = not an array, 1 = []T, 2 = [][]T, etc.
	IsPointer    bool        `json:"isPointer,omitempty"`
	PointerDepth int         `json:"pointerDepth,omitempty"` // 0 = not a pointer, 1 = *T, 2 = **T, etc.
	ElementType  string      `json:"elementType,omitempty"`  // string representation of element type (hint)
	KeyType      string      `json:"keyType,omitempty"`      // key type for map types
	ValueType    string      `json:"valueType,omitempty"`    // value type for map types
	Fields       []FieldInfo `json:"fields,omitempty"`       // nested fields if this is a struct type
}

type MethodInfo struct {
	Name   string `json:"name"`
	Params string `json:"params"` // name of the type
	Result string `json:"result"` // name of the type
}

type ServerSpec struct {
	Methods []MethodInfo        `json:"methods"`
	Types   map[string]TypeInfo `json:"types"`
}

// GetMethodSpecs returns information about all registered RPC methods.
// This can be used to generate API specifications
//
// Example usage:
//
//	spec := server.GetMethodSpecs()
//	for _, method := range spec.Methods {
//	    fmt.Printf("Method: %s\n", method.Name)
//	    fmt.Printf("Params: %s\n", method.Params)
//	    fmt.Printf("Result: %s\n", method.Result)
//	}
//	for typeName, typeInfo := range spec.Types {
//	    fmt.Printf("Type: %s\n", typeName)
//	    fmt.Printf("Fields: %+v\n", typeInfo.Fields)
//	}
func (s *Server) GetMethodSpecs() ServerSpec {
	types := make(map[string]TypeInfo)
	var methods []MethodInfo

	s.methods.Range(func(key, value interface{}) bool {
		methodName := key.(string)
		handler := value.(methodHandler)

		fnType := handler.fnValue.Type()

		paramType := fnType.In(1)
		resultType := fnType.Out(0)

		collectStructTypes(paramType, types)
		collectStructTypes(resultType, types)

		paramInfo := extractTypeInfo(paramType)
		resultInfo := extractTypeInfo(resultType)

		method := MethodInfo{
			Name:   methodName,
			Params: buildFullTypeName(paramInfo),
			Result: buildFullTypeName(resultInfo),
		}

		methods = append(methods, method)
		return true
	})

	return ServerSpec{
		Methods: methods,
		Types:   types,
	}
}

// strips all pointer levels from a type and returns the base type.
func stripPointers(typ reflect.Type) reflect.Type {
	for typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	return typ
}

// strips all pointer and array/slice levels to get the base type.
func stripArraysAndPointers(typ reflect.Type) reflect.Type {
	typ = stripPointers(typ)

	for typ.Kind() == reflect.Array || typ.Kind() == reflect.Slice {
		typ = typ.Elem()
		typ = stripPointers(typ)
	}
	return typ
}

func countPointerDepth(typ reflect.Type) (depth int, baseType reflect.Type) {
	baseType = typ
	for baseType.Kind() == reflect.Ptr {
		depth++
		baseType = baseType.Elem()
	}
	return depth, baseType
}

func countArrayDepth(typ reflect.Type) (depth int, baseType reflect.Type) {
	baseType = typ
	for baseType.Kind() == reflect.Array || baseType.Kind() == reflect.Slice {
		depth++
		baseType = baseType.Elem()
	}
	return depth, baseType
}

func stripToBaseType(typ reflect.Type) reflect.Type {
	_, baseType := countArrayDepth(typ)
	return stripPointers(baseType)
}

func collectStructTypes(typ reflect.Type, types map[string]TypeInfo) {
	typ = stripArraysAndPointers(typ)

	if typ.Kind() == reflect.Map {
		keyType := stripArraysAndPointers(typ.Key())
		valueType := stripArraysAndPointers(typ.Elem())
		collectStructTypes(keyType, types)
		collectStructTypes(valueType, types)
		return
	}

	if typ.Name() != "" && typ.PkgPath() != "" {
		unmarshalKind := getUnmarshalKind(typ)
		if unmarshalKind != "" {
			typeName := getQualifiedTypeName(typ)
			if _, exists := types[typeName]; !exists {
				info := extractTypeInfo(typ)
				types[typeName] = info
			}
		}
	}

	if typ.Kind() == reflect.Struct {
		typeName := getQualifiedTypeName(typ)

		if _, exists := types[typeName]; exists {
			return
		}

		info := extractTypeInfo(typ)
		types[typeName] = info

		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			if !field.IsExported() {
				continue
			}

			jsonTag := field.Tag.Get("json")
			if jsonTag == "-" {
				continue
			}

			fieldType := stripArraysAndPointers(field.Type)

			if fieldType.Kind() == reflect.Struct {
				collectStructTypes(fieldType, types)
			} else if fieldType.Kind() == reflect.Map {
				keyType := stripArraysAndPointers(fieldType.Key())
				valueType := stripArraysAndPointers(fieldType.Elem())
				collectStructTypes(keyType, types)
				collectStructTypes(valueType, types)
			} else {
				if fieldType.Name() != "" && fieldType.PkgPath() != "" {
					unmarshalKind := getUnmarshalKind(fieldType)
					if unmarshalKind != "" {
						collectStructTypes(fieldType, types)
					}
				}
			}
		}
	}
}

func extractTypeInfo(typ reflect.Type) TypeInfo {
	info := TypeInfo{}

	pointerDepth, baseType := countPointerDepth(typ)
	if pointerDepth > 0 {
		info.IsPointer = true
	}
	info.PointerDepth = pointerDepth
	typ = baseType

	originalTyp := typ
	arrayDepth, baseType := countArrayDepth(typ)
	if arrayDepth > 0 {
		info.IsArray = true
	}
	info.ArrayDepth = arrayDepth
	typ = baseType

	if typ.Kind() == reflect.Map {
		keyType := typ.Key()
		valueType := typ.Elem()
		info.KeyType = getQualifiedTypeNameForField(keyType)
		info.ValueType = getQualifiedTypeNameForField(valueType)
		info.Name = fmt.Sprintf("map[%s]%s", info.KeyType, info.ValueType)
		info.Kind = typ.Kind().String()

		if arrayDepth > 0 {
			elemTypeForString := originalTyp.Elem()
			info.ElementType = buildElementTypeString(elemTypeForString)
		}

		return info
	}

	if arrayDepth > 0 {
		elemTypeForString := originalTyp.Elem()
		info.ElementType = buildElementTypeString(elemTypeForString)

		typ = stripToBaseType(originalTyp)
	}

	if typ.Kind() == reflect.Struct {
		if typ.Name() != "" {
			info.Name = typ.Name()
			info.Package = typ.PkgPath()
		} else {
			info.Name = generateAnonymousStructName(typ)
		}
	} else {
		if typ.Name() != "" && typ.PkgPath() != "" {
			info.Name = typ.Name()
			info.Package = typ.PkgPath()
		} else {
			info.Name = getQualifiedTypeName(typ)
		}
	}

	unmarshalKind := getUnmarshalKind(typ)
	if unmarshalKind != "" {
		info.Kind = unmarshalKind
	} else {
		info.Kind = typ.Kind().String()
	}

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

		pointerDepth, baseType := countPointerDepth(fieldType)
		if pointerDepth > 0 {
			fieldInfo.IsPointer = true
		}
		fieldInfo.PointerDepth = pointerDepth
		fieldType = baseType

		originalFieldType := fieldType
		arrayDepth, baseType := countArrayDepth(fieldType)
		if arrayDepth > 0 {
			fieldInfo.IsArray = true
		}
		fieldInfo.ArrayDepth = arrayDepth
		fieldType = baseType

		if arrayDepth > 0 {
			elemTypeForString := originalFieldType.Elem()
			fieldInfo.ElementType = buildElementTypeString(elemTypeForString)

			fieldType = stripToBaseType(originalFieldType)
		}

		if fieldType.Kind() == reflect.Map {
			keyType := fieldType.Key()
			valueType := fieldType.Elem()
			fieldInfo.KeyType = getQualifiedTypeNameForField(keyType)
			fieldInfo.ValueType = getQualifiedTypeNameForField(valueType)
			fieldInfo.Type = fmt.Sprintf("map[%s]%s", fieldInfo.KeyType, fieldInfo.ValueType)
			fieldInfo.Kind = fieldType.Kind().String()
			fields = append(fields, fieldInfo)
			continue
		}

		fieldInfo.Type = getQualifiedTypeNameForField(fieldType)

		unmarshalKind := getUnmarshalKind(fieldType)
		if unmarshalKind != "" {
			fieldInfo.Kind = unmarshalKind
		} else {
			fieldInfo.Kind = fieldType.Kind().String()
		}

		fields = append(fields, fieldInfo)
	}

	return fields
}

func buildElementTypeString(typ reflect.Type) string {
	arrayDepth, baseType := countArrayDepth(typ)
	pointerDepth, baseType := countPointerDepth(baseType)
	baseName := getQualifiedTypeNameForField(baseType)

	for i := 0; i < pointerDepth; i++ {
		baseName = "*" + baseName
	}

	for i := 0; i < arrayDepth; i++ {
		baseName = "[]" + baseName
	}

	return baseName
}

func buildFullTypeName(info TypeInfo) string {
	name := info.Name

	if info.Kind == "map" {
		mapName := fmt.Sprintf("map[%s]%s", info.KeyType, info.ValueType)
		if info.IsArray && info.ArrayDepth > 0 {
			arrayPrefix := ""
			for i := 0; i < info.ArrayDepth; i++ {
				arrayPrefix += "[]"
			}
			mapName = arrayPrefix + mapName
		}
		for i := 0; i < info.PointerDepth; i++ {
			mapName = "*" + mapName
		}
		return mapName
	}

	if info.Package != "" {
		name = fmt.Sprintf("%s.%s", info.Package, info.Name)
	}

	if info.IsArray && info.ArrayDepth > 0 {
		arrayPrefix := ""
		for i := 0; i < info.ArrayDepth; i++ {
			arrayPrefix += "[]"
		}
		if info.ElementType != "" {
			name = arrayPrefix + info.ElementType
		} else {
			name = arrayPrefix + name
		}
	} else {
		for i := 0; i < info.PointerDepth; i++ {
			name = "*" + name
		}
	}

	return name
}

func getUnmarshalKind(typ reflect.Type) string {
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	method, found := typ.MethodByName("UnmarshalKind")
	if found {
		if method.Type.NumOut() == 1 && method.Type.Out(0).Kind() == reflect.String {
			zeroValue := reflect.New(typ).Elem()
			results := method.Func.Call([]reflect.Value{zeroValue})
			if len(results) > 0 {
				return results[0].String()
			}
		}
	}

	ptrType := reflect.PointerTo(typ)
	method, found = ptrType.MethodByName("UnmarshalKind")
	if found {
		if method.Type.NumOut() == 1 && method.Type.Out(0).Kind() == reflect.String {
			zeroValuePtr := reflect.New(typ)
			results := method.Func.Call([]reflect.Value{zeroValuePtr})
			if len(results) > 0 {
				return results[0].String()
			}
		}
	}

	return ""
}

func generateAnonymousStructName(typ reflect.Type) string {
	typeString := typ.String()
	hash := sha256.Sum256([]byte(typeString))
	hashStr := hex.EncodeToString(hash[:])
	return fmt.Sprintf("struct_%s", hashStr[:16])
}

func getQualifiedTypeNameForField(typ reflect.Type) string {
	qualifiedName := getQualifiedTypeName(typ)
	if qualifiedName == "struct" && typ.Kind() == reflect.Struct {
		return generateAnonymousStructName(typ)
	}
	return qualifiedName
}

func getQualifiedTypeName(typ reflect.Type) string {
	if typ.Name() != "" && typ.PkgPath() != "" {
		return fmt.Sprintf("%s.%s", typ.PkgPath(), typ.Name())
	}

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
		if typ.Name() == "" {
			return generateAnonymousStructName(typ)
		}
		pkgPath := typ.PkgPath()
		if pkgPath != "" {
			return fmt.Sprintf("%s.%s", pkgPath, typ.Name())
		}
		return typ.Name()
	case reflect.Array, reflect.Slice:
		return fmt.Sprintf("[]%s", getQualifiedTypeName(typ.Elem()))
	case reflect.Map:
		return fmt.Sprintf("map[%s]%s", getQualifiedTypeName(typ.Key()), getQualifiedTypeName(typ.Elem()))
	default:
		if typ.Name() != "" {
			pkgPath := typ.PkgPath()
			if pkgPath != "" {
				return fmt.Sprintf("%s.%s", pkgPath, typ.Name())
			}
			return typ.Name()
		}
		return typ.String()
	}
}

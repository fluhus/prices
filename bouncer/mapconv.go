package bouncer

// Convertion of int maps to string maps and back, for serializability.

import (
	"fmt"
	"reflect"
	"strconv"
)

// Converts a map with int keys to a map with string keys.
func intMapToStringMap(m interface{}) interface{} {
	// Extract types.
	mapType := reflect.TypeOf(m)
	if mapType.Kind() != reflect.Map {
		panic("Expected map, got " + mapType.Kind().String() + ".")
	}
	keyType := mapType.Key()
	if keyType.Kind() != reflect.Int {
		panic("Expected int key, got " + keyType.Kind().String() + ".")
	}
	valType := mapType.Elem()

	// Create result.
	resultType := reflect.MapOf(reflect.TypeOf(""), valType)
	result := reflect.MakeMap(resultType)
	mapValue := reflect.ValueOf(m)
	for _, key := range mapValue.MapKeys() {
		str := reflect.ValueOf(fmt.Sprint(key.Int()))
		result.SetMapIndex(str, mapValue.MapIndex(key))
	}

	return result.Interface()
}

// Converts a map with string keys to a map with int keys.
func stringMapToIntMap(m interface{}) interface{} {
	// Extract types.
	mapType := reflect.TypeOf(m)
	if mapType.Kind() != reflect.Map {
		panic("Expected map, got " + mapType.Kind().String() + ".")
	}
	keyType := mapType.Key()
	if keyType.Kind() != reflect.String {
		panic("Expected string key, got " + keyType.Kind().String() + ".")
	}
	valType := mapType.Elem()

	// Create result.
	resultType := reflect.MapOf(reflect.TypeOf(0), valType)
	result := reflect.MakeMap(resultType)
	mapValue := reflect.ValueOf(m)
	for _, key := range mapValue.MapKeys() {
		i, err := strconv.Atoi(key.String())
		if err != nil {
			panic("Failed to convert to int: " + err.Error())
		}
		result.SetMapIndex(reflect.ValueOf(i), mapValue.MapIndex(key))
	}

	return result.Interface()
}

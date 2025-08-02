package utils

import "reflect"

// IsZero 判断任意类型的值是否为零值或空值
func IsZero(v interface{}) bool {
	val := reflect.ValueOf(v)
	// 处理指针类型：如果指针为nil，或指向的值为零值，则返回true
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return true
		}
		val = val.Elem() // 获取指针指向的具体值
	}

	// 根据具体类型判断零值
	switch val.Kind() {
	case reflect.String:
		return val.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return val.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return val.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return val.Float() == 0
	case reflect.Bool:
		return !val.Bool() // 如果为false，则视为零值
	case reflect.Struct:
		// 结构体：递归检查所有字段是否为零值
		for i := 0; i < val.NumField(); i++ {
			if !IsZero(val.Field(i).Interface()) {
				return false
			}
		}
		return true
	default:
		// 其他类型（如Slice、Map等）直接与零值比较
		return reflect.DeepEqual(val.Interface(), reflect.Zero(val.Type()).Interface())
	}
}

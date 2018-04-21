package convert

import 	"strconv"

func ToIntFromString(str string) (int, error) {
	new, err := strconv.Atoi(str)
	return new, err
}

func ToInt32FromString(str string) (int32, error) {
	i64, err := strconv.ParseInt(str, 10, 32)
	i32 := int32(i64)
	return i32, err
	
}

func ToInt64FromString(str string) (int64, error) {
	i, err := strconv.ParseInt(str, 10, 64)
	return i, err
}

func ToFloat64FromString(str string) (float64, error) {
	f, err := strconv.ParseFloat(str, 64)
	return f, err
}

func ToBoolFromString(str string) (bool, error) {
	b, err := strconv.ParseBool(str)
	return b, err
}

func ToStringFromInt64(value int64) string {
	return strconv.FormatInt(value, 10)
}

func ToStringFromFloat64(value float64) string {
	return strconv.FormatFloat(value, 'G', -1, 64)
}

func ToStringFromBool(value bool) string {
	return strconv.FormatBool(value)
}

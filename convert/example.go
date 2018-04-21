package main

import (
	"fmt"
	"convert"
	"reflect"
)

func main() {
	
	var number int = 3290423909
		
	string := convert.ToStringFromInt64(int64(number))

	fmt.Println(string, reflect.TypeOf(string)) // "3290423909" string

	var num int

	num, _ = convert.ToIntFromString(string)

	fmt.Println(num, reflect.TypeOf(num))	// 3290423909 int
	
}

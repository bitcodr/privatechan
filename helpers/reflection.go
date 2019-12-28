package helpers

import "reflect"

import "fmt"

//Invoke function
func Invoke(any interface{}, name string, args ...interface{}) {
	inputs := make([]reflect.Value, len(args))
	for i := 0; i < len(args); i++ {
		inputs[i] = reflect.ValueOf(args[i])
	}
	fmt.Println(reflect.TypeOf(any))
	reflect.ValueOf(any).MethodByName(name).Call(inputs)
}

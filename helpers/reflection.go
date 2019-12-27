package helpers

import "reflect"

//Invoke function
func Invoke(any interface{}, name string, args... interface{}) bool {
	inputs := make([]reflect.Value, len(args))
	for i, _ := range args {
		inputs = append(inputs, reflect.ValueOf(args[i]))
	}
	return reflect.ValueOf(any).MethodByName(name).Call(inputs)[0].Bool()
}

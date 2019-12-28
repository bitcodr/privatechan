package helpers

import "reflect"

//Invoke function
func Invoke(any interface{}, name string, args ...interface{}) bool {
	inputs := make([]reflect.Value, len(args))
	for i := 0; i < len(args); i++ {
		inputs[i] = reflect.ValueOf(args[i])
	}
	return reflect.ValueOf(any).MethodByName(name).Call(inputs)[0].Bool()
}

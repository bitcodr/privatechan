package helpers

import "reflect"

//Invoke function
func Invoke(any interface{}, result *bool, name string, args ...interface{}) {
	inputs := make([]reflect.Value, len(args))
	for i := 0; i < len(args); i++ {
		inputs[i] = reflect.ValueOf(args[i])
	}
	*result = reflect.ValueOf(any).MethodByName(name).Call(inputs)[0].Bool()
}

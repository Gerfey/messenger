package stamps

import "reflect"

type HandledStamp struct {
	Handler    string
	Result     any
	ResultType reflect.Type
}

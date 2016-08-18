package inji

import "reflect"

var _g *Graph

func InitDefault() {
	_g = NewGraph()
}

func Close() {
	_g.Close()
}

func SetLogger(logger Logger) {
	_g.Logger = logger
}

func RegisterOrFail(name string, value interface{}) {
	_g.RegisterOrFail(name, value)
}

func Register(name string, value interface{}) error {
	return _g.Register(name, value)
}

func RegisterOrFailSingle(name string, value interface{}) {
	_g.RegisterOrFailSingle(name, value)
}

func RegisterSingle(name string, value interface{}) error {
	return _g.RegisterSingle(name, value)
}

func FindByType(t reflect.Type) (interface{}, bool) {
	o, ok := _g.FindByType(t)
	if !ok || o == nil || o.Value == nil {
		return nil, false
	}
	return o.Value, ok
}

func Find(name string) (interface{}, bool) {
	o, ok := _g.Find(name)
	if !ok || o == nil || o.Value == nil {
		return nil, false
	}
	return o.Value, ok
}

func GraphLen() int {
	return _g.Len()
}

func GraphPrint() string {
	return _g.SPrint()
}

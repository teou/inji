/**
1.if singleton, every struct ptr inject object can only have one
instance in memory, we set twice in named to
ensure that(first by name, then by struct ptr type)

2.TODO every struct inject object(not a ptr) should be handled like:
create new struct on every inject
	`inject:""` by type,must be a pointer to a struct
	`inject:"devService"`

3.singleton(default false)
	`singleton:"true"`
	`singleton:"false"`

4.cannil(default false)
	`cannil:"true"`
	`cannil:"false"`
**/
package inji

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/facebookgo/structtag"
	"github.com/teou/ordered_map"
)

type Logger interface {
	Debug(format interface{}, v ...interface{})
	Info(format interface{}, v ...interface{})
	Error(format interface{}, v ...interface{}) error
}

type Object struct {
	Name        string
	reflectType reflect.Type
	Value       interface{}
	closed      bool
}

func (o Object) String() string {
	if o.reflectType.Kind() == reflect.Ptr {
		return fmt.Sprintf(`{"name":"%s","type":"%v","value":"%p"}`, o.Name, o.reflectType, o.Value)
	} else {
		return fmt.Sprintf(`{"name":"%s","type":"%v"}`, o.Name, o.reflectType)
	}
}

type Graph struct {
	l      sync.RWMutex
	Logger Logger
	named  *ordered_map.OrderedMap
}

func NewGraph() *Graph {
	g := &Graph{
		named: ordered_map.NewOrderedMap(),
	}
	return g
}

func getTypeName(t reflect.Type) string {
	isptr := false
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		isptr = true
	}
	pkg := t.PkgPath()
	name := ""
	if pkg != "" {
		name = pkg + "." + t.Name()
	} else {
		name = t.Name()
	}
	if isptr {
		name = "*" + name
	}
	return name
}

func (g *Graph) FindByType(t reflect.Type) (*Object, bool) {
	g.l.RLock()
	defer g.l.RUnlock()
	return g.findByType(t)
}

func (g *Graph) findByType(t reflect.Type) (*Object, bool) {
	n := getTypeName(t)
	return g.find(n)
}

func (g *Graph) Len() int {
	g.l.RLock()
	defer g.l.RUnlock()
	return g._len()
}

func (g *Graph) _len() int {
	return g.named.Len()
}

func (g *Graph) Find(name string) (*Object, bool) {
	g.l.RLock()
	defer g.l.RUnlock()
	return g.find(name)
}

func (g *Graph) find(name string) (*Object, bool) {
	f, ok := g.named.Get(name)
	if !ok {
		return nil, false
	}
	ret, ok := f.(*Object)
	if !ok {
		g.named.Delete(name)
		return nil, false
	} else {
		return ret, true
	}
}

func (g *Graph) del(name string) {
	g.named.Delete(name)
}

func (g *Graph) set(name string, o *Object) {
	g.named.Set(name, o)
}

func (g *Graph) setboth(name string, o *Object) {
	g.named.Set(name, o)
	if isStructPtr(o.reflectType) {
		tn := getTypeName(o.reflectType)
		g.named.Set(tn, o)
	}
}

func (g *Graph) RegisterOrFailNoFill(name string, value interface{}) interface{} {
	v, err := g.RegisterNoFill(name, value)
	if err != nil {
		if g.Logger != nil {
			g.Logger.Error(err)
		}
		panic(err.Error())
	}
	return v
}

func (g *Graph) RegisterOrFailSingleNoFill(name string, value interface{}) interface{} {
	v, err := g.RegisterSingleNoFill(name, value)
	if err != nil {
		if g.Logger != nil {
			g.Logger.Error(err)
		}
		panic(err.Error())
	}
	return v
}

func (g *Graph) RegisterOrFail(name string, value interface{}) interface{} {
	v, err := g.Register(name, value)
	if err != nil {
		if g.Logger != nil {
			g.Logger.Error(err)
		}
		panic(err.Error())
	}
	return v
}

func (g *Graph) RegisterOrFailSingle(name string, value interface{}) interface{} {
	v, err := g.RegisterSingle(name, value)
	if err != nil {
		if g.Logger != nil {
			g.Logger.Error(err)
		}
		panic(err.Error())
	}
	return v
}

func (g *Graph) RegisterNoFill(name string, value interface{}) (interface{}, error) {
	g.l.Lock()
	defer g.l.Unlock()
	return g.register(name, value, false, true)
}

func (g *Graph) RegisterSingleNoFill(name string, value interface{}) (interface{}, error) {
	g.l.Lock()
	defer g.l.Unlock()
	return g.register(name, value, true, true)
}

func (g *Graph) Register(name string, value interface{}) (interface{}, error) {
	g.l.Lock()
	defer g.l.Unlock()
	return g.register(name, value, false, false)
}

func (g *Graph) RegisterSingle(name string, value interface{}) (interface{}, error) {
	g.l.Lock()
	defer g.l.Unlock()
	return g.register(name, value, true, false)
}

func (g *Graph) register(name string, value interface{}, singleton bool, noFill bool) (interface{}, error) {
	reflectType := reflect.TypeOf(value)

	if isStructPtr(reflectType) {
		if name == "" {
			name = getTypeName(reflectType)
		}
	} else {
		if name == "" {
			return nil, fmt.Errorf("name can not be empty,name=%s,type=%v", name, reflectType)
		}
	}

	//already registered
	found, ok := g.find(name)
	if ok {
		return nil, fmt.Errorf("already registered,name=%s,type=%v,found=%v", name, reflectType, found)
	}

	o := &Object{
		Name:        name,
		reflectType: reflectType,
	}
	if isStructPtr(o.reflectType) {
		t := reflectType.Elem()
		var v reflect.Value
		created := false
		if isNil(value) {
			created = true
			v = reflect.New(t)
		} else {
			v = reflect.ValueOf(value)
		}

		for i := 0; i < t.NumField(); i++ {
			if !created && noFill {
				continue
			}

			f := t.Field(i)
			vfe := v.Elem()
			vf := vfe.Field(i)

			if vf.CanInterface() {
				if !isZeroOfUnderlyingType(vf.Interface()) {
					continue
				}
			}

			ok, tag, err := structtag.Extract("inject", string(f.Tag))
			if err != nil {
				return nil, fmt.Errorf("extract tag fail,f=%s,err=%v", f.Name, err)
			}
			if !ok {
				continue
			}

			if f.Anonymous || !vf.CanSet() {
				return nil, fmt.Errorf("inject tag must on a public field!field=%s,type=%s", f.Name, t.Name())
				continue
			}

			_, singletonStr, _ := structtag.Extract("singleton", string(f.Tag))
			singletonTag := false
			if singletonStr == "true" {
				singletonTag = true
			}
			_, canNilStr, _ := structtag.Extract("cannil", string(f.Tag))
			canNil := false
			if canNilStr == "true" {
				canNil = true
			}

			var found *Object
			if tag != "" {
				//due to default singleton of struct ptr injections
				//we should first find by name,then find by type
				found, ok = g.find(tag)
				if singletonTag && !ok && isStructPtr(f.Type) {
					found, ok = g.findByType(f.Type)
				}
			} else {
				found, ok = g.findByType(f.Type)
			}

			if !ok || found == nil {
				if canNil {
					continue
				}
				if isStructPtr(f.Type) {
					_, err := g.register(tag, reflect.NewAt(f.Type.Elem(), nil).Interface(), singletonTag, noFill)
					if err != nil {
						return nil, err
					}
				} else {
					return nil, fmt.Errorf("dependency field=%s,tag=%s not found in object %s:%v", f.Name, tag, name, reflectType)
				}

				if tag != "" {
					found, ok = g.find(tag)
					if !ok && singleton {
						found, ok = g.findByType(f.Type)
					}
				} else {
					found, ok = g.findByType(f.Type)
				}
			}

			if !ok || found == nil {
				return nil, fmt.Errorf("dependency %s not found in object %s:%v", f.Name, name, reflectType)
			}

			reflectFoundValue := reflect.ValueOf(found.Value)
			if !found.reflectType.AssignableTo(f.Type) {
				switch reflectFoundValue.Kind() {
				case reflect.Int:
					fallthrough
				case reflect.Int8:
					fallthrough
				case reflect.Int16:
					fallthrough
				case reflect.Int32:
					fallthrough
				case reflect.Int64:
					iv := reflectFoundValue.Int()
					switch f.Type.Kind() {
					case reflect.Int:
						fallthrough
					case reflect.Int8:
						fallthrough
					case reflect.Int16:
						fallthrough
					case reflect.Int32:
						fallthrough
					case reflect.Int64:
						vf.SetInt(iv)
					default:
						return nil, fmt.Errorf("dependency name=%s,type=%v not valid in object %s:%v", f.Name, f.Type, name, reflectType)
					}
				case reflect.Float32:
					fallthrough
				case reflect.Float64:
					fv := reflectFoundValue.Float()
					switch f.Type.Kind() {
					case reflect.Float32:
						fallthrough
					case reflect.Float64:
						vf.SetFloat(fv)
					default:
						return nil, fmt.Errorf("dependency name=%s,type=%v not valid in object %s:%v", f.Name, f.Type, name, reflectType)
					}
				default:
					return nil, fmt.Errorf("dependency name=%s,type=%v not valid in object %s:%v", f.Name, f.Type, name, reflectType)
				}
			} else {
				vf.Set(reflectFoundValue)
			}
		}
		o.Value = v.Interface()
	} else {
		//TODO
		//if inejection type is a struct(not a pointer),
		//we should create a new struct every time
		//when a inject tag is found,and no *Object is created and
		//the created struct should NOT be set into
		//graph.named(or there will be a memory leak)!
		//same as a bean's prototype scope in spring.
		//otherwise all inject dependency will behave like
		//spring's singleton bean scope
		if canNil(value) && isNil(value) {
			return nil, fmt.Errorf("register nil on name=%s, val=%v", name, value)
		}
		o.Value = value
	}

	//depedency resolved, init the object
	canStart, ok := o.Value.(Startable)
	if ok {
		err := canStart.Start()
		if err != nil {
			return nil, err
		}
	}

	//set to graph
	if isStructPtr(reflectType) && singleton {
		g.setboth(name, o)
	} else {
		g.set(name, o)
	}
	if g.Logger != nil {
		g.Logger.Debug("registered!name=%s,t=%v,v=%v", name, reflectType, o.Value)
	}
	return o.Value, nil
}

func (g *Graph) SPrint() string {
	g.l.RLock()
	defer g.l.RUnlock()
	return g.sPrint()
}

func (g *Graph) sPrint() string {
	ret := "["
	iter := g.named.IterFunc()
	count := g._len()
	i := 0
	for kv, ok := iter(); ok; kv, ok = iter() {
		str := fmt.Sprintf(`{"key":"%s","object":%s}`, fmt.Sprintf("%s", kv.Key), fmt.Sprintf("%s", kv.Value))

		ret = ret + str
		i++
		if i < count {
			ret = ret + ","
		}
	}
	ret = ret + "]"
	return ret
}

//beaware of the close order when use g.Close!
//every *Object will be Closed on reverse order
//of the Register
//there should be no defer xx.Close betwen g.Register
//function calls in main.exe
func (g *Graph) Close() {
	g.l.Lock()
	defer g.l.Unlock()

	if g.Logger != nil {
		g.Logger.Info("close objects %v", g.sPrint())
	}
	var keys []string
	iter := g.named.RevIterFunc()
	for kv, ok := iter(); ok; kv, ok = iter() {
		k, ok := kv.Key.(string)
		if !ok {
			continue
		}
		keys = append(keys, k)
		o, ok := kv.Value.(*Object)
		if !ok {
			continue
		}
		if o.closed {
			continue
		}
		if isStructPtr(o.reflectType) {
			keys = append(keys, getTypeName(o.reflectType))
		}
		if o.Value == nil {
			continue
		}
		c, ok := o.Value.(Closeable)
		if ok {
			c.Close()
			if g.Logger != nil {
				g.Logger.Debug("closed!object=%s", o)
			}
			o.closed = true
		}
	}

	for _, k := range keys {
		g.del(k)
	}
	if g.Logger != nil {
		g.Logger.Info("inject graph closed all")
	}
}

func isStructPtr(t reflect.Type) bool {
	return t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct
}

func canNil(v interface{}) bool {
	k := reflect.ValueOf(v).Kind()
	return k == reflect.Ptr || k == reflect.Interface
}

func isNil(v interface{}) bool {
	return reflect.ValueOf(v).IsNil()
}

func isZeroOfUnderlyingType(x interface{}) bool {
	if x == nil {
		return true
	}
	rv := reflect.ValueOf(x)
	k := rv.Kind()

	if k == reflect.Func {
		return rv.IsNil()
	}

	if (k == reflect.Ptr || k == reflect.Interface || k == reflect.Chan || k == reflect.Map || k == reflect.Slice) && rv.IsNil() {
		return true
	}

	switch k {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		if rv.Len() <= 0 {
			return true
		} else {
			return false
		}
	}
	return x == reflect.Zero(reflect.TypeOf(x)).Interface()
}

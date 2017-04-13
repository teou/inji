package inji

import (
	"fmt"
	"os"
	"reflect"
	"testing"
)

type Log struct {
}

func (l *Log) Error(format interface{}, args ...interface{}) error {
	f, _ := format.(string)
	s := fmt.Sprintf(f, args...)
	fmt.Println(s)
	fmt.Fprint(os.Stderr, s)
	return nil
}

func (l *Log) Info(format interface{}, args ...interface{}) {
	f, _ := format.(string)
	s := fmt.Sprintf(f, args...)
	fmt.Println(s)
}

func (l *Log) Debug(format interface{}, args ...interface{}) {
	f, _ := format.(string)
	s := fmt.Sprintf(f, args...)
	fmt.Println(s)
}

type Test interface {
	Test()
}

type Test1 struct {
	Conf string `inject:"conf"`
	Int1 *int   `inject:"int1"`
}

type Test2 struct {
	Test1 *Test1 `inject:""`
}

type Test3 struct {
	Test2 Test `inject:"test2"`
}

func (t *Test1) Start() error {
	fmt.Println("test1.start,conf="+t.Conf, *t.Int1)
	return nil
}

func (t *Test3) Start() error {
	fmt.Println("test3.start")
	return nil
}

func (t *Test1) Close() {
	fmt.Println("test1.close")
}

func (t *Test2) Close() {
	fmt.Println("test2.close")
}

func (t *Test3) Close() {
	fmt.Println("test3.close")
}

func (t *Test1) Test() {
	fmt.Println("test1.test,conf="+t.Conf, *t.Int1)
}

func (t *Test2) Test() {
	t.Test1.Test()
	fmt.Println("test2.test")
}

func (t *Test3) Test() {
	t.Test2.Test()
	fmt.Println("test3.test")
}

func TestFindByName(t *testing.T) {
	InitDefault()
	defer Close()
	_, err := Register("test2", (*Test2)(nil))
	if err == nil {
		t.Error("conf is not registered need error")
		return
	}

	i1 := 123
	_, err = Register("int1", &i1)
	if err != nil {
		t.Error(err)
		return
	}

	_, err = Register("conf", "##conf1")
	if err != nil {
		t.Error(err)
		return
	}

	var t2tmp interface{}
	t2tmp, err = Register("test2", (*Test2)(nil))
	if err != nil {
		t.Error(err)
		return
	}
	t2t, ok := t2tmp.(*Test2)
	if !ok || t2t == nil {
		t.Error("invalid t2t %v", t2tmp)
		return
	}

	_, err = Register("test3", (*Test3)(nil))
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println("###############")
	fmt.Println("init finished,len=", GraphLen())
	fmt.Println("###############")
	fmt.Println("graph=", GraphPrint())
	fmt.Println("###############")

	t1o, ok := FindByType(reflect.TypeOf((*Test1)(nil)))
	if !ok || t1o == nil {
		t.Error("t1 not found", t1o, ok)
		return
	}
	t2o, ok := Find("test2")
	if !ok || t2o == nil {
		t.Error("t2 not found", t2o, ok)
		return
	}
	t3o, ok := Find("test3")
	if !ok || t3o == nil {
		t.Error("t3 not found", t3o, ok)
		return
	}
	t1, _ := t1o.(*Test1)
	t2, _ := t2o.(*Test2)
	t3, _ := t3o.(*Test3)

	fmt.Println("#################")
	t1.Test()
	fmt.Println("#################")
	t2.Test()
	fmt.Println("#################")
	t3.Test()
	fmt.Println("#################")
}

type Sin1 struct {
	Name string
}

func (s *Sin1) Start() error {
	fmt.Println("start", s.Name)
	return nil
}

func (s *Sin1) Close() {
	fmt.Println("close", s.Name)
}

type Sin2 struct {
	Sin1 *Sin1  `inject:"sin1" singleton:"false" cannil:"false"`
	Sin2 *Sin1  `inject:"sin2" singleton:"false"`
	Sin3 *Sin1  `inject:"sin3" singleton:"true"`
	Sin4 int    `inject:"sin4" cannil:"true"`
	Sin5 string `inject:"sin5" cannil:"true"`
	Sin6 string `inject:"sin6"`
}

func TestNoneSingleton(t *testing.T) {
	fmt.Println("######  singleton test")
	InitDefault()
	l := &Log{}
	SetLogger(l)
	defer Close()

	s1 := &Sin1{Name: "s1"}
	RegisterOrFailSingle("sin1", s1)
	s2 := &Sin1{Name: "s2"}
	RegisterOrFail("sin2", s2)
	RegisterOrFail("sin6", "haha sin6")

	RegisterOrFail("s", (*Sin2)(nil))

	so, ok := Find("s")
	if !ok {
		t.Error("s not found")
		return
	}
	s, ok := so.(*Sin2)
	if !ok {
		t.Error("s not valid")
		return
	}
	if s.Sin1.Name != "s1" || s.Sin2.Name != "s2" {
		t.Error("injection not valid", s)
	}
	fmt.Println("s1.name", s.Sin1.Name)
	fmt.Println("s2.name", s.Sin2.Name)
}

func TestCanNil(t *testing.T) {
	fmt.Println("######  cannil test")
	InitDefault()
	l := &Log{}
	SetLogger(l)
	defer Close()

	s2 := &Sin1{Name: "s2"}
	RegisterOrFail("sin2", s2)

	_, err := Register("s", (*Sin2)(nil))
	if err == nil {
		t.Error("sin6 not registered should fail!", err)
		return
	} else {
		fmt.Println("sin6 not registered:)", err)
	}

	RegisterOrFail("sin6", "haha sin6")
	RegisterOrFail("s", (*Sin2)(nil))

	so, ok := Find("s")
	if !ok {
		t.Error("s not found")
		return
	}
	s, ok := so.(*Sin2)
	if !ok {
		t.Error("s not valid")
		return
	}
	if s.Sin1.Name != "" || s.Sin2.Name != "s2" {
		t.Error("injection not valid", s)
		return
	}
	if s.Sin3 == nil {
		t.Error("sin3 is nil")
		return
	} else {
		fmt.Println("s.Sin3", s.Sin3)
	}

	s3f, ok := FindByType(reflect.TypeOf((*Sin1)(nil)))
	if !ok {
		t.Error("s3 should be singleton")
		return
	}
	s3o, ok := s3f.(*Sin1)
	if !ok {
		t.Error("s3 type not valid")
		return
	}
	if s3o != s.Sin3 {
		t.Error("singleton not the same", s3o, s.Sin3)
		return
	}

	if s.Sin4 != 0 {
		t.Error("sin4 must be 0")
		return
	}
	if s.Sin5 != "" {
		t.Error("sin5 must be empty string")
		return
	}

}

type IA1 struct {
	A int64   `inject:"a"`
	B float32 `inject:"b"`
}

func TestIntAssign(t *testing.T) {
	fmt.Println("############## test assign")
	InitDefault()
	l := &Log{}
	SetLogger(l)
	defer Close()

	var ta int
	ta = 126
	RegisterOrFail("a", ta)
	var tb float64
	tb = 32.11134
	RegisterOrFail("b", tb)
	RegisterOrFail("c", (*IA1)(nil))
	ia1o, ok := Find("c")
	if !ok {
		t.Error("c not found")
		return
	}
	ia1, ok := ia1o.(*IA1)
	if !ok {
		t.Error("c invalid")
		return
	}
	if ia1.A != 126 {
		t.Error("ia1.A invalid %d", ia1.A)
	} else {
		fmt.Println("ia1.A", ia1.A)
	}
	if ia1.B != 32.11134 {
		t.Error("ia1.B invalid %f", ia1.B)
	} else {
		fmt.Println("ia1.B", ia1.B)
	}

}

func TestIntAssign2(t *testing.T) {
	fmt.Println("############## test assign2")
	InitDefault()
	l := &Log{}
	SetLogger(l)
	defer Close()

	var ta int
	ta = 126
	RegisterOrFail("a", ta)
	var tb float64
	tb = 32.11134
	RegisterOrFail("b", tb)
	c := &IA1{
		A: 999,
		B: 11.22,
	}
	RegisterOrFail("c", c)
	ia1o, ok := Find("c")
	if !ok {
		t.Error("c not found")
		return
	}
	ia1, ok := ia1o.(*IA1)
	if !ok {
		t.Error("c invalid")
		return
	}
	if ia1.A != 999 {
		t.Error("ia1.A invalid %d", ia1.A)
	} else {
		fmt.Println("ia1.A", ia1.A)
	}
	if ia1.B != 11.22 {
		t.Error("ia1.B invalid %f", ia1.B)
	} else {
		fmt.Println("ia1.B", ia1.B)
	}

}

type Rec struct {
	HaHa string `inject:"haha"`
}

type Rec1 struct {
	Rec  *Rec  `inject:"rec"`
	Rec2 *Rec2 `inject:"rec2"`
}

type Rec2 struct {
	Rec  *Rec  `inject:"rec"`
	Rec1 *Rec1 `inject:"rec1"`
}

//causing a stack overflow error
func _TestRec(t *testing.T) {
	fmt.Println("############## test rec")
	InitDefault()
	l := &Log{}
	SetLogger(l)
	defer Close()

	defer func() {
		if x := recover(); x != nil {
			fmt.Println("panic : ", x)
		}
	}()
	haha := "aa"
	RegisterOrFail("haha", haha)

	//this will cause:
	//runtime: goroutine stack exceeds 1000000000-byte limit
	//fatal error: stack overflow
	RegisterOrFail("rec1", (*Rec1)(nil))
	RegisterOrFail("rec2", (*Rec2)(nil))
}

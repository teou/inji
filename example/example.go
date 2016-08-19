package main

import (
	"fmt"

	"github.com/teou/inji"
)

type Te struct {
	Target int `inject:"target"`
}

func (t *Te) Start() error {
	fmt.Println("start", t.Target)
	return nil
}

func (t *Te) Close() {
	fmt.Println("close", t.Target)
}

type Dep struct {
	Test *Te `inject:"test"`
}

func (d *Dep) Close() {
	fmt.Println("close Dep", d.Test)
}

func main() {
	inji.InitDefault()
	//dep.Close, test.Close will be called orderly
	defer inji.Close()
	inji.RegisterOrFail("target", 123)
	//test will be auto created, test.Start will be called, then dep.Start(if any)
	inji.RegisterOrFail("dep", (*Dep)(nil))
	test, _ := inji.Find("test")
	fmt.Println("find test", test)
	dep, _ := inji.Find("dep")
	fmt.Println("find dep", dep)
}

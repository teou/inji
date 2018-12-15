package main

import (
	"fmt"

	"github.com/teou/inji"
)

type Test struct {
	Target int `inject:"target"`
}

func (t *Test) Start() error {
	fmt.Println("start", t.Target)
	return nil
}

func (t *Test) Close() {
	fmt.Println("close", t.Target)
}

type Dep struct {
	Test *Test `inject:"test"`
}

func (d *Dep) Close() {
	fmt.Println("close Dep", d.Test)
}

func main() {
	inji.InitDefault()
	//dep.Close, test.Close will be called orderly
	defer inji.Close()
	inji.Reg("target", 123)
	//test will be auto created, test.Start will be called, then dep.Start(if any)
	dep := inji.Reg("dep", (*Dep)(nil)).(*Dep)
	fmt.Println("find dep", dep)
}

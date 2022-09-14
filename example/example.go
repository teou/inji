package main

import (
	"fmt"
	"runtime/debug"

	"github.com/teou/inji"
)

type Test struct {
	Target      int            `inject:"target"`
	Timeout     int            `inject:"timeout"`
	PathString  string         `inject:"path_string"`
	PathStrings []string       `inject:"path_strings"`
	PathMap     map[string]int `inject:"path_map"`
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
	Wait int   `inject:"wait"`
}

func (d *Dep) Close() {
	fmt.Println("close Dep", d.Test)
}

func main() {
	defer func() {
		if e := recover(); e != nil {
			fmt.Println("panic:", e)
			fmt.Println(string(debug.Stack()))
		}
	}()

	inji.InitDefault()
	//dep.Close, test.Close will be called orderly
	defer inji.Close()
	inji.Reg("target", 123)
	inji.Reg("wait", 123)
	inji.Reg("timeout", 123)
	inji.Reg("path_string", "path string")
	inji.Reg("path_strings", []string{"path1", "path2"})
	inji.Reg("path_map", map[string]int{"path1": 1, "path2": 2})

	//test will be auto created, test.Start will be called, then dep.Start(if any)
	dep := inji.Reg("dep", (*Dep)(nil)).(*Dep)
	fmt.Println("find dep", dep)

	fmt.Println(inji.GraphPrint())

	fmt.Println(inji.GraphPrintTree())
}

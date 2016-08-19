# inji

    inji is a dependency injection container for golang.
    it auto store and register objects into a graph.
    struct pointer dependency will be auto created if 
    not found in graph.
    when closing graph, every object will be closed on a 
    reverse order of their creation.
    
# use

```go

import (
	"fmt"
	"miot/inji"
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

func example() {
	inji.InitDefault()
	//test.Close, dep.Close will be called orderly
	defer inji.Close()
	inji.RegisterOrFail("target", 123)
	//test will be auto created, test.Start will be called, then dep.Start
	inji.RegisterOrFail("dep", (*Dep)(nil))
	test, _ := inji.Find("test")
	fmt.Println("find test", test)
	dep, _ := inji.Find("dep")
	fmt.Println("find dep", dep)
}

```

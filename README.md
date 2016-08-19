# inji

    inji is a dependency injection container for golang.
    it auto store and register objects into a graph.
    struct pointer dependency will be auto created if 
    not found in graph.
    when closing graph, every object will be closed on a 
    reverse order of their creation.
    
# use

```go
type Test struct {
    Target int `inject:"target"`
}
func (t *Test) Start() error {
    fmt.Println("start", t.Target)
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

inji.InitDefault()
//test.Close, dep.Close will be called orderly
defer inji.Close() 
inji.RegisterOrFail("target", 123)
//test will be auto created, test.Start will be called, then dep.Start
inji.RegisterOrFail("dep",(*Dep)(nil)) 
test, ok := inji.Find("test")
dep, ok := inji.Find("dep")
```

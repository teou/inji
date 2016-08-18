# inji
    inji is a dependency injection container for golang.
    it auto store and register objects into a graph.
    struct pointer dependency will be auto created if 
    not found in graph.
    when closing graph, every object will be closed on a 
    reverse order of their creation.
# use
    inji.InitDefault()
    defer inji.Close()
    inji.RegisterOrFail("target", 123)
    target, ok := inji.Find("target")

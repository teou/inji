package inji

type Startable interface {
	Start() error
}

type Closeable interface {
	Close()
}

type Injectable interface {
	Startable
	Closeable
}

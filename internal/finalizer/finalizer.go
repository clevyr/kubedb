package finalizer

import "sync"

//nolint:gochecknoglobals
var Default = &Finalizers{}

type Finalizers struct {
	finalizers []func(err error)
	mu         sync.RWMutex
}

func (f *Finalizers) Add(fn ...func(err error)) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.finalizers = append(f.finalizers, fn...)
}

func Add(fn ...func(err error)) {
	Default.Add(fn...)
}

func (f *Finalizers) PostRun(err error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	for _, fn := range f.finalizers {
		fn(err)
	}
}

func PostRun(err error) {
	Default.PostRun(err)
}

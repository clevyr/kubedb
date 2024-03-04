package cmd

var finalizers []func(err error)

func OnFinalize(y ...func(err error)) {
	finalizers = append(finalizers, y...)
}

func PostRun(err error) {
	for _, x := range finalizers {
		x(err)
	}
}

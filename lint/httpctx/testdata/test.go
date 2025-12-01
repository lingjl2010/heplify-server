package testdata

import (
	"os/exec"
	"time"
)

type A struct {
}

func (a A) Deadline() (deadline time.Time, ok bool) {
	//TODO implement me
	panic("implement me")
}

func (a A) Done() <-chan struct{} {
	//TODO implement me
	panic("implement me")
}

func (a A) Err() error {
	//TODO implement me
	panic("implement me")
}

func (a A) Value(key interface{}) interface{} {
	//TODO implement me
	panic("implement me")
}

func Test() {
	exec.CommandContext(A{}, "") // want `shouldn't pass .+ as context.Context`
}

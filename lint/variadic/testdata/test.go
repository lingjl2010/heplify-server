package testdata

func a(f ...string) {
}

func b(f ...any) {
}

func c(f ...interface{}) {

}

type any2 any

func d(f ...any2) {}

func e(f ...func()) {}

func f(f ...*string) {}

func Test() {
	a("a", "b", "c")
	b(nil) // want `the last params is nil, maybe misuse`
	b(1, 2, 3)
	c(nil) // want `the last params is nil, maybe misuse`
	c(1, 2, 3)
	d(nil) // want `the last params is nil, maybe misuse`
	d(1, 2, 3)
	e(nil) // want `the last params is nil, maybe misuse`
	e(func() {})
	f(nil)           // want `the last params is nil, maybe misuse`
	f(nil, nil, nil) // want `the last params is nil, maybe misuse`
	_ = append([]any{}, nil)
}

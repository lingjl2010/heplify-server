package testdata

import "io" // want "should import package errors"

func a() error {
	return nil
}

var errFoo = io.EOF

func Test() {
	var err, err2 error
	if err == io.EOF { // want "direct comparison of errors"
	}
	if err != io.EOF { // want "direct comparison of errors"
	}
	if err != nil {
	}
	if err != err2 { // want "direct comparison of two complex errors"
	}
	if err != a() { // want "direct comparison of two complex errors"
	}
	if io.EOF != a() { // want "direct comparison of errors"
	}
	if a() != io.EOF { // want "direct comparison of errors"
	}
	if io.EOF == err { // want "direct comparison of errors"
	}
	if err != errFoo { // want "direct comparison of errors"
	}
	switch err { // want "direct comparison of errors"
	case io.EOF:
	}
}

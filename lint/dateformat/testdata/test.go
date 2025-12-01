// Package testdata
package testdata

import (
	"fmt"
	"time"
)

func a() {
	time.Now().Format("2006-01-02 15:04:05") // want `should use time.DateTime`

	fmt.Println("2006-01-02 15:04:05") // want `should use time.DateTime`
}

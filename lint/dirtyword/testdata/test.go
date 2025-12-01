// Package testdata
package testdata

// import comment
import "fmt"

// 坐席 独立注释 // want `shouldn't use `

// 坐席 测试注释 // want `shouldn't use `
type 坐席 struct { // want `shouldn't use `
}

// 坐席分组 // want `shouldn't use `
func a() {
	// 函数内坐席  // want `shouldn't use `
	fmt.Println("坐席") // want `shouldn't use `
	fmt.Println("replace   into abc") // want `shouldn't use .*replace.*into`
}

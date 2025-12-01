package testdata

import (
	"context"
	"database/sql"
	"os"
	"os/exec"
	"os/signal"
)

func F()                         {}
func FContext(_ context.Context) {}

type DB struct {
	*sql.DB
}

func (db *DB) CloseContext(_ context.Context) {}

type DB2 = DB

// Generices function
//
//	func Generices[T any](_ context.Context, data T) T {
//		return data
//	}

type PhoneNumberData interface {
	Number() string
}

func buildDataMap[T PhoneNumberData](ctx context.Context, dataList []T) map[string]T {
	dataMap := make(map[string]T)
	for i, data := range dataList {
		if _, ok := dataMap[data.Number()]; ok {
			// log.Warn(ctx, fmt.Sprintf("duplicate number %s", data.Number()), otel2.Any("data", data))
			continue
		}
		dataMap[data.Number()] = dataList[i]
	}
	return dataMap
}

func GF[T1, T2 any](v1 T1, v2 T2) {}

func GFContext[T1, T2 any](_ context.Context, v1 T1, v2 T2) {}

func A() {
	a := sql.DB{}
	// method of object
	_, _ = a.Exec("") // want `ExecContext should be used instead`

	b := DB{}
	// method of embedded struct
	_ = b.Close() // want `CloseContext should be used instead`

	// method of embedded struct
	_ = b.Ping() // want `PingContext should be used instead`

	// method of literal object
	_ = (&sql.DB{}).Ping() // want `PingContext should be used instead`

	b2 := DB2{}
	_ = b2.Ping() // want `PingContext should be used instead`

	_ = b2.Close()

	// static method of other pkg
	exec.Command("") // want `CommandContext should be used instead`

	// static method of self pkg
	F() // want `FContext should be used instead`

	// static method of self pkg
	GF[uint64, uint64](1, 2) // want `GFContext should be used instead`

	// allow list
	signal.Notify(nil, os.Interrupt)
}

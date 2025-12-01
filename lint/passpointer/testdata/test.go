package testdata

import (
	"database/sql"
	"fmt"
)

func testSQL() {
	var a sql.Row
	var v1, v2, v3 int
	_ = a.Scan(v1, &v2, v3) // want `call of Scan\(\[0 2]\) passes non-pointer`

	var va []any
	// 变参暂时检查不了
	_ = a.Scan(va...)

	_, _ = fmt.Sscan("1 2 3", v1, &v2, v3) // want `call of Sscan\(\[1 3]\) passes non-pointer`
}

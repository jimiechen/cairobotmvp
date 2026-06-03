package sqlite

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const DriverName = "sqlite"

type Dialector = gorm.Dialector

func Open(dataSourceName string) Dialector {
	return sqlite.Open(dataSourceName)
}

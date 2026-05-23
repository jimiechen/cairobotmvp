package repository

import "testing"

func TestMySQLRepo_Interface(t *testing.T) {
	var _ I18nRepository = (*MySQLRepo)(nil)
}

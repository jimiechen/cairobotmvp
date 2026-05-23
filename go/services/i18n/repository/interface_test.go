package repository

import "testing"

func TestI18nRepository_Interface(t *testing.T) {
	var _ I18nRepository = (*SQLiteRepo)(nil)
}

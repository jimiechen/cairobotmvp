package service

import "testing"

func TestI18nService_Interface(t *testing.T) {
	var _ I18nService = (*I18nServiceImpl)(nil)
}

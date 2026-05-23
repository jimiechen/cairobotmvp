package sdk

import "errors"

var (
	ErrServiceRequired = errors.New("config service is required in InProcess mode")
	ErrUnsupportedMode = errors.New("unsupported SDK mode")
	ErrModuleNotFound  = errors.New("module not found")
	ErrFieldNotFound   = errors.New("field not found")
	ErrTypeMismatch    = errors.New("field type mismatch")
	ErrBindFailed      = errors.New("bind to struct failed")
)

module github.com/jimiechen/mineplanet/go/services/i18n

go 1.23

require (
	github.com/jimiechen/mineplanet/go/common-lib v0.0.0
	github.com/jimiechen/mineplanet/protocols/generated/go/base v0.0.0
)

replace (
	github.com/jimiechen/mineplanet/go/common-lib => ../../common-lib
	github.com/jimiechen/mineplanet/protocols/generated/go/base => ../../../proto/generated/go/base
)

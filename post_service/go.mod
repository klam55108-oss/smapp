module smapp/post_service

go 1.22.6

replace smapp/common => ../common

require (
	github.com/go-ozzo/ozzo-validation/v4 v4.3.0
	github.com/go-sql-driver/mysql v1.8.1
	github.com/google/uuid v1.6.0
	smapp/common v0.0.0-00010101000000-000000000000
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/asaskevich/govalidator v0.0.0-20200108200545-475eaeb16496 // indirect
)

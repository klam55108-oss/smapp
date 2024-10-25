module smapp/image

go 1.22.6

replace smapp/common => ../common

require (
	google.golang.org/grpc v1.67.1
	smapp/common v0.0.0-00010101000000-000000000000
)

require (
	golang.org/x/net v0.28.0 // indirect
	golang.org/x/sys v0.24.0 // indirect
	golang.org/x/text v0.17.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240814211410-ddb44dafa142 // indirect
	google.golang.org/protobuf v1.35.1 // indirect
)

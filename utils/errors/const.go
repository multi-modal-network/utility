package errors

const (
	CodeInner                  = 10000
	CodeSetupDBFailed          = 10001
	CodeUpdateTopoFailed       = 10002
	CodeTopoUpdateDeviceFailed = 10003
	CodeTopoUpdateLinkFailed   = 10004
)

var (
	Unknown                = New(CodeInner, "unknown error")
	SetupDBFailed          = New(CodeSetupDBFailed, "service setup db failed")
	UpdateTopoFailed       = New(CodeUpdateTopoFailed, "update topo failed")
	TopoUpdateDeviceFailed = New(CodeTopoUpdateDeviceFailed, "update device failed")
	TopoUpdateLinkFailed   = New(CodeTopoUpdateLinkFailed, "update link failed")
)

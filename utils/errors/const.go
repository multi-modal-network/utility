package errors

const (
	CodeInner                  = 10000
	CodeSetupDBFailed          = 10001
	CodeUpdateTopoFailed       = 10002
	CodeTopoUpdateDeviceFailed = 10003
	CodeTopoUpdateLinkFailed   = 10004
	CodeDataBaseConfigEmpty    = 10005
	CodeRegisterDatabaseFailed = 10006
	CodeRunSyncdbFailed        = 10007
	CodeGetTofinoPortFailed    = 10008
	CodePrepareFlowFailed      = 10009
)

var (
	Unknown                = New(CodeInner, "unknown error")
	SetupDBFailed          = New(CodeSetupDBFailed, "service setup db failed")
	UpdateTopoFailed       = New(CodeUpdateTopoFailed, "update topo failed")
	TopoUpdateDeviceFailed = New(CodeTopoUpdateDeviceFailed, "update device failed")
	TopoUpdateLinkFailed   = New(CodeTopoUpdateLinkFailed, "update link failed")
	DataBaseConfigEmpty    = New(CodeDataBaseConfigEmpty, "database config empty")
	RegisterDatabaseFailed = New(CodeRegisterDatabaseFailed, "register database failed")
	RunSyncdbFailed        = New(CodeRunSyncdbFailed, "run syncdb failed")
	GetTofinoPortFailed    = New(CodeGetTofinoPortFailed, "get tofino port failed")
	PrepareFlowFailed      = New(CodePrepareFlowFailed, "prepare flow failed")
)

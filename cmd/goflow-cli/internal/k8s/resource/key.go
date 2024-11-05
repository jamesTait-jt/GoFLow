package resource

type Key int

const (
	Namespace Key = iota
	MessageBrokerDeployment
	MessageBrokerService
	GRPCServerDeployment
	GRPCServerService
	WorkerpoolDeployment
	WorkerpoolPV
	WorkerpoolPVC
)

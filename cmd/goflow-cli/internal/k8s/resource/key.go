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

func (k Key) String() string {
	switch k {
	case Namespace:
		return "namespace"

	case MessageBrokerDeployment:
		return "messageBrokerDeployment"

	case MessageBrokerService:
		return "messageBrokerService"

	case GRPCServerDeployment:
		return "gRPCServerDeployment"

	case GRPCServerService:
		return "gRPCServerService"

	case WorkerpoolDeployment:
		return "workerpoolDeployment"

	case WorkerpoolPV:
		return "workerpoolPV"

	case WorkerpoolPVC:
		return "workerpoolPVC"

	default:
		return ""
	}
}

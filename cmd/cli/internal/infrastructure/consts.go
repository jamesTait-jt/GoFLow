package infrastructure

var (
	RedisContainerName = "goflow-redis-container"

	GRPCServerContainerName = "goflow-grpc-container"

	// GRPCContainerPort is the port that the gRPC server will listen on inside the container
	GRPCContainerPort int32 = 50051

	// WorkerpoolHandlersLocation is the location of the handlers on the pluginBuilder and workerpool containers
	WorkerpoolHandlersLocation = "/app/handlers"
)

package analyzer

type ProjectAnalysisResult struct {
	resources                       []Resource
	resourceToResourceUsageBindings []ResourceToResourceUsageBinding
	projectToResourceMappings       []ProjectToResourceMapping
}

type Resource struct {
	resourceName string
	resourceType ResourceType
}

type ResourceType string

const (
	AzureDatabaseForPostgresql ResourceType = "azure.db.postgresql"
	AzureDatabaseForMysql      ResourceType = "azure.db.mysql"
	AzureCacheForRedis         ResourceType = "azure.db.redis"
	AzureCosmosDBForMongoDB    ResourceType = "azure.db.cosmos.mongo"
	AzureCosmosDBForNoSQL      ResourceType = "azure.db.cosmos.nosql"
	AzureContainerApp          ResourceType = "azure.host.containerapp"
	AzureOpenAiModel           ResourceType = "azure.ai.openai.model"
	AzureServiceBus            ResourceType = "azure.messaging.servicebus"
	AzureEventHubs             ResourceType = "azure.messaging.eventhubs"
	AzureStorageAccount        ResourceType = "azure.storage"
)

type ResourceToResourceUsageBinding struct {
	sourceResourceName string
	targetResourceName string
}

type ProjectToResourceMapping struct {
	projectRelativePath string
	resourceName        string
}

func mergeProject(result1 ProjectAnalysisResult, result2 ProjectAnalysisResult) ProjectAnalysisResult {
	// todo: handle duplicated error
	return ProjectAnalysisResult{
		resources: append(result1.resources, result2.resources...),
		resourceToResourceUsageBindings: append(result1.resourceToResourceUsageBindings,
			result2.resourceToResourceUsageBindings...),
		projectToResourceMappings: append(result1.projectToResourceMappings, result2.projectToResourceMappings...),
	}
}

package analyzer

type ProjectAnalysisResult struct {
	Resources                       []Resource
	ResourceToResourceUsageBindings []ResourceToResourceUsageBinding
	ProjectToResourceMappings       []ProjectToResourceMapping
}

type Resource struct {
	ResourceName string
	ResourceType ResourceType
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
	SourceResourceName string
	TargetResourceName string
}

type ProjectToResourceMapping struct {
	ProjectRelativePath string
	ResourceName        string
}

func mergeProject(result1 ProjectAnalysisResult, result2 ProjectAnalysisResult) ProjectAnalysisResult {
	// todo: handle duplicated error
	return ProjectAnalysisResult{
		Resources: append(result1.Resources, result2.Resources...),
		ResourceToResourceUsageBindings: append(result1.ResourceToResourceUsageBindings,
			result2.ResourceToResourceUsageBindings...),
		ProjectToResourceMappings: append(result1.ProjectToResourceMappings, result2.ProjectToResourceMappings...),
	}
}

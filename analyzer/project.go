package analyzer

type Project struct {
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

func mergeProject(project1 Project, project2 Project) Project {
	// todo: handle duplicated error
	return Project{
		resources: append(project1.resources, project2.resources...),
		resourceToResourceUsageBindings: append(project1.resourceToResourceUsageBindings,
			project2.resourceToResourceUsageBindings...),
		projectToResourceMappings: append(project1.projectToResourceMappings, project2.projectToResourceMappings...),
	}
}

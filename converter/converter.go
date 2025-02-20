package converter

import (
	"fmt"

	"ajpa/analyzer"
	"ajpa/converter/azd"
)

func ProjectAnalysisResultToAzdProjectConfig(result analyzer.ProjectAnalysisResult) (azd.ProjectConfig, error) {
	config := azd.ProjectConfig{
		Name: result.Name,
	}
	config.Services = make(map[string]*azd.ServiceConfig)
	for name, app := range result.Applications {
		config.Services[name] = &azd.ServiceConfig{
			Name:         name,
			RelativePath: app.ProjectRelativePath,
			Host:         azd.ContainerAppTarget, // todo: support other kinds.
			Language:     azd.ServiceLanguageJava,
		}
	}
	config.Resources = make(map[string]*azd.ResourceConfig)
	for name, service := range result.Services {
		resourceType, err := toResourceType(service)
		if err != nil {
			return azd.ProjectConfig{}, err
		}
		props, err := toProps(service)
		if err != nil {
			return azd.ProjectConfig{}, err
		}
		config.Resources[name] = &azd.ResourceConfig{
			Name:  name,
			Type:  resourceType,
			Props: props,
		}
	}
	for appName, serviceNameMap := range result.ApplicationToBackingService {
		hostingName := result.ApplicationToHostingService[appName]
		for serviceName := range serviceNameMap {
			config.Resources[hostingName].Uses = append(config.Resources[hostingName].Uses, serviceName)
		}
	}
	return config, nil
}

func toProps(service analyzer.Service) (interface{}, error) {
	switch s := service.(type) {
	case analyzer.AzureContainerApp:
		return azd.ContainerAppProps{
			Port: 8080, // todo: support non-web app.
		}, nil
	case analyzer.AzureDatabaseForPostgresql, // todo: Add database name in PostgresqlProps
		analyzer.AzureDatabaseForMysql,
		analyzer.AzureCacheForRedis,
		analyzer.AzureCosmosDbForMongoDb,
		analyzer.AzureCosmosDb:
		return nil, nil
	case analyzer.AzureServiceBus:
		return azd.ServiceBusProps{
			Queues: s.Queues,
			Topics: s.Topics,
		}, nil
	case analyzer.AzureEventHubs:
		return azd.EventHubsProps{
			Hubs: s.Hubs,
		}, nil
	case analyzer.AzureStorageAccount:
		return azd.StorageProps{
			Containers: s.Containers,
		}, nil
	default:
		return "", fmt.Errorf("unknown service type when get Props: %v", service)
	}
}

func toResourceType(service analyzer.Service) (azd.ResourceType, error) {
	switch service.(type) {
	case analyzer.AzureContainerApp:
		return azd.ResourceTypeHostContainerApp, nil
	case analyzer.AzureDatabaseForPostgresql:
		return azd.ResourceTypeDbPostgres, nil
	case analyzer.AzureDatabaseForMysql:
		return azd.ResourceTypeDbPostgres, nil // todo: change to mysql when azd support mysql
	case analyzer.AzureCacheForRedis:
		return azd.ResourceTypeDbRedis, nil
	case analyzer.AzureCosmosDbForMongoDb:
		return azd.ResourceTypeDbMongo, nil
	case analyzer.AzureCosmosDb:
		return azd.ResourceTypeDbMongo, nil // todo: change to cosmos when azd support cosmos
	case analyzer.AzureServiceBus:
		return azd.ResourceTypeMessagingServiceBus, nil
	case analyzer.AzureEventHubs:
		return azd.ResourceTypeMessagingEventHubs, nil
	case analyzer.AzureStorageAccount:
		return azd.ResourceTypeStorage, nil
	default:
		return "", fmt.Errorf("unknown service type: %v", service)
	}
}

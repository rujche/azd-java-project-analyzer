package analyzer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"ajpa/analyzer/internal"
)

func AnalyzeJavaProject(projectRootPath string) (ProjectAnalysisResult, error) {
	return analyzeJavaProjectSubDirectory(projectRootPath, projectRootPath)
}

func analyzeJavaProjectSubDirectory(projectRootPath string, subDirectoryPath string) (ProjectAnalysisResult, error) {
	entries, err := os.ReadDir(subDirectoryPath)
	if err != nil {
		return ProjectAnalysisResult{}, fmt.Errorf("reading directory: %w", err)
	}
	result := ProjectAnalysisResult{
		Name: filepath.Base(projectRootPath),
	}
	for _, entry := range entries {
		if entry.IsDir() {
			newResult, err := analyzeJavaProjectSubDirectory(projectRootPath,
				filepath.Join(subDirectoryPath, entry.Name()))
			if err != nil {
				return ProjectAnalysisResult{}, fmt.Errorf("analyzing java project: %w", err)
			}
			result, err = mergeProjectAnalysisResult(result, newResult)
			if err != nil {
				return ProjectAnalysisResult{}, err
			}
		} else {
			// todo:
			// 1. Support file names like backend-pom.xml
			// 2. Support build.gradle
			if strings.ToLower(entry.Name()) == "pom.xml" {
				pomPath := filepath.Join(subDirectoryPath, entry.Name())
				newResult, err := analyzePomProject(projectRootPath, pomPath)
				if err != nil {
					return ProjectAnalysisResult{}, err
				}
				// todo: consider multiple pom use same Azure resource
				result, err = mergeProjectAnalysisResult(result, newResult)
				if err != nil {
					return ProjectAnalysisResult{}, err
				}
			}
		}
	}
	return result, nil
}

func analyzePomProject(projectRootPath string, pomFileAbsolutePath string) (ProjectAnalysisResult, error) {
	pom, err := internal.CreateEffectivePom(pomFileAbsolutePath)
	if err != nil {
		return ProjectAnalysisResult{}, fmt.Errorf("creating effective pom: %w", err)
	}
	pomRelativePathPath, err := filepath.Rel(projectRootPath, pomFileAbsolutePath)
	if err != nil {
		return ProjectAnalysisResult{}, err
	}
	pom.PomFilePath = pomRelativePathPath
	if !isSpringBootRunnableProject(pom) {
		return ProjectAnalysisResult{}, nil
	}
	result := ProjectAnalysisResult{}
	projectRelativePath := filepath.Dir(pomRelativePathPath)
	// 1. Add Application
	applicationName := internal.GetNameFromDirPath(filepath.Dir(pomFileAbsolutePath))
	err = addApplicationToResult(&result, applicationName, Application{projectRelativePath})
	if err != nil {
		return result, err
	}
	// 2. Add Application related hosting Service
	hostingServiceName := applicationName
	err = addApplicationRelatedHostingServiceToResult(&result, applicationName, hostingServiceName, AzureContainerApp{})
	if err != nil {
		return result, err
	}
	// 3. Add Application related backing Service
	properties := internal.ReadProperties(filepath.Dir(pomFileAbsolutePath))
	if err = detectPostgresql(&result, applicationName, pom, properties); err != nil {
		return ProjectAnalysisResult{}, err
	}
	if err = detectMysql(&result, applicationName, pom, properties); err != nil {
		return ProjectAnalysisResult{}, err
	}
	if err = detectRedis(&result, applicationName, pom); err != nil {
		return ProjectAnalysisResult{}, err
	}
	if err = detectMongo(&result, applicationName, pom); err != nil {
		return ProjectAnalysisResult{}, err
	}
	if err = detectCosmos(&result, applicationName, pom); err != nil {
		return ProjectAnalysisResult{}, err
	}
	if err = detectServiceBus(&result, applicationName, pom, properties); err != nil {
		return ProjectAnalysisResult{}, err
	}
	if err = detectEventHubs(&result, applicationName, pom, properties); err != nil {
		return ProjectAnalysisResult{}, err
	}
	if err = detectStorageAccount(&result, applicationName, pom, properties); err != nil {
		return ProjectAnalysisResult{}, err
	}
	return result, nil
}

func isSpringBootRunnableProject(pom internal.Pom) bool {
	if len(pom.Modules) > 0 {
		return false
	}
	for _, dep := range pom.Build.Plugins {
		if dep.GroupId == "org.springframework.boot" && dep.ArtifactId == "spring-boot-maven-plugin" {
			return true
		}
	}
	return false
}

func detectPostgresql(result *ProjectAnalysisResult, applicationName string, pom internal.Pom,
	properties map[string]string) error {
	if hasDependency(pom, "org.postgresql", "postgresql") ||
		hasDependency(pom, "com.azure.spring", "spring-cloud-azure-starter-jdbc-postgresql") {
		return addApplicationRelatedBackingServiceToResult(result, applicationName, DefaultPostgresqlServiceName,
			AzureDatabaseForPostgresql{getDatabaseNameFromSpringDataSourceUrlProperty(properties)})
	}
	return nil
}

func detectMysql(result *ProjectAnalysisResult, applicationName string, pom internal.Pom,
	properties map[string]string) error {
	if hasDependency(pom, "com.mysql", "mysql-connector-j") ||
		hasDependency(pom, "com.azure.spring", "spring-cloud-azure-starter-jdbc-mysql") {
		// todo:
		// 1. support multiple container app use multiple mysql
		// 2. Support multiple container app use one mysql
		// 3. Same to other resources like postgresql
		return addApplicationRelatedBackingServiceToResult(result, applicationName, DefaultMysqlServiceName,
			AzureDatabaseForMysql{getDatabaseNameFromSpringDataSourceUrlProperty(properties)})
	}
	return nil
}

func detectRedis(result *ProjectAnalysisResult, applicationName string, pom internal.Pom) error {
	if hasDependency(pom, "org.springframework.boot", "spring-boot-starter-data-redis") ||
		hasDependency(pom, "org.springframework.boot", "spring-boot-starter-data-redis-reactive") {
		return addApplicationRelatedBackingServiceToResult(result, applicationName, DefaultRedisServiceName,
			AzureCacheForRedis{})
	}
	return nil
}

func detectMongo(result *ProjectAnalysisResult, applicationName string, pom internal.Pom) error {
	if hasDependency(pom, "org.springframework.boot", "spring-boot-starter-data-mongodb") ||
		hasDependency(pom, "org.springframework.boot", "spring-boot-starter-data-mongodb-reactive") {
		return addApplicationRelatedBackingServiceToResult(result, applicationName, DefaultMongoServiceName,
			AzureCosmosDbForMongoDb{})
	}
	return nil
}

func detectCosmos(result *ProjectAnalysisResult, applicationName string, pom internal.Pom) error {
	if hasDependency(pom, "com.azure.spring", "spring-cloud-azure-starter-data-cosmos") {
		return addApplicationRelatedBackingServiceToResult(result, applicationName, DefaultCosmosServiceName,
			AzureCosmosDb{})
	}
	return nil
}

func detectServiceBus(result *ProjectAnalysisResult, applicationName string, pom internal.Pom,
	properties map[string]string) error {
	if !hasDependency(pom, "com.azure.spring", "spring-cloud-azure-starter-servicebus-jms") &&
		!hasDependency(pom, "com.azure.spring", "spring-cloud-azure-stream-binder-servicebus") {
		return nil
	}
	var queues []string
	if hasDependency(pom, "com.azure.spring", "spring-cloud-azure-stream-binder-servicebus") {
		queues = internal.AppendAndDistinct(queues, internal.GetDistinctBindingDestinationValues(properties))
	}
	return addApplicationRelatedBackingServiceToResult(result, applicationName, DefaultServiceBusServiceName,
		AzureServiceBus{Queues: queues})
}

func detectEventHubs(result *ProjectAnalysisResult, applicationName string, pom internal.Pom,
	properties map[string]string) error {
	if !hasDependency(pom, "com.azure.spring", "spring-cloud-azure-stream-binder-eventhubs") &&
		!hasDependency(pom, "com.azure.spring", "spring-cloud-azure-starter-eventhubs") &&
		!hasDependency(pom, "com.azure.spring", "spring-cloud-azure-starter-integration-eventhubs") &&
		!hasDependency(pom, "com.azure.spring", "spring-messaging-azure-eventhubs") &&
		!hasDependency(pom, "com.azure.spring", "spring-cloud-starter-stream-kafka") &&
		!hasDependency(pom, "com.azure.spring", "spring-kafka") {
		return nil
	}
	var hubs []string
	if hasDependency(pom, "com.azure.spring", "spring-cloud-azure-stream-binder-eventhubs") ||
		hasDependency(pom, "org.springframework.cloud", "spring-cloud-starter-stream-kafka") {
		hubs = internal.AppendAndDistinct(hubs, internal.GetDistinctBindingDestinationValues(properties))
	}
	if hasDependency(pom, "com.azure.spring", "spring-cloud-azure-starter-eventhubs") {
		var targetPropertyNames = []string{
			"spring.cloud.azure.eventhubs.event-hub-name",
			"spring.cloud.azure.eventhubs.producer.event-hub-name",
			"spring.cloud.azure.eventhubs.consumer.event-hub-name",
			"spring.cloud.azure.eventhubs.processor.event-hub-name",
		}
		eventHubsNamePropertyMap := map[string]string{}
		for _, propertyName := range targetPropertyNames {
			if propertyValue, ok := properties[propertyName]; ok {
				eventHubsNamePropertyMap[propertyName] = propertyValue
			}
		}
		hubs = internal.AppendAndDistinct(hubs, internal.DistinctMapValues(eventHubsNamePropertyMap))
	}
	return addApplicationRelatedBackingServiceToResult(result, applicationName, DefaultEventHubsServiceName,
		AzureEventHubs{Hubs: hubs})
}

func detectStorageAccount(result *ProjectAnalysisResult, applicationName string, pom internal.Pom,
	properties map[string]string) error {
	if !hasDependency(pom, "com.azure.spring", "spring-cloud-azure-stream-binder-eventhubs") &&
		!hasDependency(pom, "com.azure.spring", "spring-cloud-azure-starter-integration-eventhubs") &&
		!hasDependency(pom, "com.azure.spring", "spring-messaging-azure-eventhubs") {
		return nil
	}
	var containers []string
	if (hasDependency(pom, "com.azure.spring", "spring-cloud-azure-stream-binder-eventhubs") &&
		containsInKeywordInBindingName(properties)) ||
		hasDependency(pom, "com.azure.spring", "spring-cloud-azure-starter-integration-eventhubs") ||
		hasDependency(pom, "com.azure.spring", "spring-messaging-azure-eventhubs") {
		containerNamePropertyMap := make(map[string]string)
		for key, value := range properties {
			if strings.HasSuffix(key, "spring.cloud.azure.eventhubs.processor.checkpoint-store.container-name") {
				containerNamePropertyMap[key] = value
			}
		}
		containers = internal.AppendAndDistinct(containers, internal.DistinctMapValues(containerNamePropertyMap))
	}
	return addApplicationRelatedBackingServiceToResult(result, applicationName, DefaultStorageServiceName,
		AzureStorageAccount{Containers: containers})
}

func hasDependency(pom internal.Pom, groupId string, artifactId string) bool {
	for _, dep := range pom.Dependencies {
		if dep.GroupId == groupId && dep.ArtifactId == artifactId {
			return true
		}
	}
	return false
}

func getDatabaseNameFromSpringDataSourceUrlProperty(properties map[string]string) string {
	databaseName := ""
	databaseNamePropertyValue, ok := properties["spring.datasource.url"]
	if ok {
		databaseName = internal.GetDatabaseName(databaseNamePropertyValue)
	}
	return databaseName
}

func containsInKeywordInBindingName(properties map[string]string) bool {
	bindingDestinations := internal.GetBindingDestinationMap(properties)
	for bindingName := range bindingDestinations {
		if strings.Contains(bindingName, "-in-") { // Example: consume-in-0
			return true
		}
	}
	return false
}

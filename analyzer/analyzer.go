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
	bindingDestinationMap := internal.GetBindingDestinationMap(properties)
	bindingDestinationValues := internal.DistinctValues(bindingDestinationMap)
	for _, dep := range pom.Dependencies {
		if dep.GroupId == "com.azure.spring" && dep.ArtifactId == "spring-cloud-azure-starter-servicebus-jms" {
			err = addApplicationRelatedBackingServiceToResult(&result, applicationName, DefaultServiceBusServiceName,
				AzureServiceBus{})
			if err != nil {
				return result, err
			}
		} else if dep.GroupId == "com.azure.spring" && dep.ArtifactId == "spring-cloud-azure-stream-binder-servicebus" {
			err = addApplicationRelatedBackingServiceToResult(&result, applicationName, DefaultServiceBusServiceName,
				AzureServiceBus{Queues: bindingDestinationValues})
			if err != nil {
				return result, err
			}
		} else if dep.GroupId == "com.azure.spring" && dep.ArtifactId == "spring-cloud-azure-stream-binder-eventhubs" {
			err = addApplicationRelatedBackingServiceToResult(&result, applicationName, DefaultEventHubsServiceName,
				AzureEventHubs{Hubs: bindingDestinationValues})
			if err != nil {
				return result, err
			}
		} else if dep.GroupId == "com.azure.spring" && dep.ArtifactId == "spring-cloud-azure-starter-eventhubs" {
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
			eventHubsNamePropertyValues := internal.DistinctValues(eventHubsNamePropertyMap)
			err = addApplicationRelatedBackingServiceToResult(&result, applicationName, DefaultEventHubsServiceName,
				AzureEventHubs{Hubs: eventHubsNamePropertyValues})
			if err != nil {
				return result, err
			}
		} else if dep.GroupId == "com.azure.spring" && dep.ArtifactId == "spring-cloud-azure-starter-integration-eventhubs" {
			// eventhubs name is empty here because no configured property
			err = addApplicationRelatedBackingServiceToResult(&result, applicationName, DefaultEventHubsServiceName,
				AzureEventHubs{})
			if err != nil {
				return result, err
			}
		} else if dep.GroupId == "com.azure.spring" && dep.ArtifactId == "spring-messaging-azure-eventhubs" {
			// eventhubs name is empty here because no configured property
			err = addApplicationRelatedBackingServiceToResult(&result, applicationName, DefaultEventHubsServiceName,
				AzureEventHubs{})
			if err != nil {
				return result, err
			}
		} else if dep.GroupId == "org.springframework.cloud" && dep.ArtifactId == "spring-cloud-starter-stream-kafka" {
			// todo: 1. add spring boot version related property. 2. Differentiate event hub and event hub kafka.
			err = addApplicationRelatedBackingServiceToResult(&result, applicationName, DefaultEventHubsServiceName,
				AzureEventHubs{Hubs: bindingDestinationValues})
			if err != nil {
				return result, err
			}
		} else if dep.GroupId == "org.springframework.kafka" && dep.ArtifactId == "spring-kafka" {
			// eventhubs name is empty here because no configured property
			err = addApplicationRelatedBackingServiceToResult(&result, applicationName, DefaultEventHubsServiceName,
				AzureEventHubs{})
			if err != nil {
				return result, err
			}
		}
		// todo: support other resource types.
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

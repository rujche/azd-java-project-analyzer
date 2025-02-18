package analyzer

import "fmt"

type ProjectAnalysisResult struct {
	Name                        string
	Applications                map[string]Application            // application name -> Application
	Services                    map[string]Service                // service name -> Service
	ApplicationToHostingService map[string]string                 // application name -> hosting Service name
	ApplicationToBackingService map[string]map[string]interface{} // application name -> backing Service names (set)
}

type Application struct {
	// todo: add other fields like Dockerfile path
	ProjectRelativePath string
}

type Service interface {
}

type AzureContainerApp struct { // todo: Support other hosting Service like AKS.
}

const DefaultPostgresqlServiceName string = "postgresql"

type AzureDatabaseForPostgresql struct {
	// todo: Add fields like auth type
	DatabaseName string
}

const DefaultMysqlServiceName = "mysql"

type AzureDatabaseForMysql struct {
	// todo: Add fields like auth type
	DatabaseName string
}

const DefaultRedisServiceName = "redis"

type AzureCacheForRedis struct {
	// todo: Add fields like auth type
}

const DefaultMongoServiceName = "redis"

type AzureCosmosDbForMongoDb struct {
	// todo: Add fields like auth type
}

const DefaultServiceBusServiceName = "service-bus"

type AzureServiceBus struct {
	Queues []string
	Topics []string
}

func addApplicationToResult(result *ProjectAnalysisResult, applicationName string, application Application) error {
	if _, ok := result.Applications[applicationName]; ok {
		return fmt.Errorf("applicationName %s already exists", applicationName)
	}
	if result.Applications == nil {
		result.Applications = make(map[string]Application)
	}
	result.Applications[applicationName] = application
	return nil
}

func addApplicationRelatedHostingServiceToResult(result *ProjectAnalysisResult, applicationName string,
	hostingServiceName string, hostingService Service) error {
	// 1. Check applicationName exists
	if _, ok := result.Applications[applicationName]; !ok {
		return fmt.Errorf("applicationName %s doesn't exist", applicationName)
	}
	// 2. Add hosting Service
	if result.Services == nil {
		result.Services = make(map[string]Service)
	}
	if _, ok := result.Services[hostingServiceName]; ok {
		return fmt.Errorf("hostingServiceName %s already exists", hostingServiceName)
	}
	result.Services[hostingServiceName] = hostingService
	// 3. Add Application to hosting Service mapping
	if result.ApplicationToHostingService == nil {
		result.ApplicationToHostingService = make(map[string]string)
	}
	if _, ok := result.ApplicationToHostingService[applicationName]; ok {
		return fmt.Errorf("applicationToHostingService (applicationName = %s) already exists", applicationName)
	}
	result.ApplicationToHostingService[applicationName] = hostingServiceName
	return nil
}

func addApplicationRelatedBackingServiceToResult(result *ProjectAnalysisResult, applicationName string,
	backingServiceName string, backingService Service) error {
	// 1. Check applicationName exists
	if _, ok := result.Applications[applicationName]; !ok {
		return fmt.Errorf("applicationName %s doesn't exist", applicationName)
	}
	// 2. Add backing Service
	if result.Services == nil {
		result.Services = make(map[string]Service)
	}
	// todo: support multiple application use same backing Service,
	// merge properties (like database name) instead of return error
	if _, ok := result.Services[backingServiceName]; ok {
		return fmt.Errorf("backingServiceName %s already exists", backingServiceName)
	}
	result.Services[backingServiceName] = backingService
	// 3. Add Application to backing Service mapping
	if result.ApplicationToBackingService == nil {
		result.ApplicationToBackingService = make(map[string]map[string]interface{})
	}
	if result.ApplicationToBackingService[applicationName] == nil {
		result.ApplicationToBackingService[applicationName] = make(map[string]interface{})
	}
	if _, ok := result.ApplicationToBackingService[applicationName][backingServiceName]; ok {
		return fmt.Errorf("applicationToBackingService (%s -> %s) already exists", applicationName, backingServiceName)
	}
	result.ApplicationToBackingService[applicationName][backingServiceName] = ""
	return nil
}

func mergeProjectAnalysisResult(result1 ProjectAnalysisResult, result2 ProjectAnalysisResult) (ProjectAnalysisResult,
	error) {
	// 1. Add application
	for applicationName, application := range result2.Applications {
		err := addApplicationToResult(&result1, applicationName, application)
		if err != nil {
			return ProjectAnalysisResult{}, err
		}
	}
	// 2. Add application hosting Service
	for applicationName, hostingServiceName := range result2.ApplicationToHostingService {
		hostingService, ok := result2.Services[hostingServiceName]
		if !ok {
			return ProjectAnalysisResult{}, fmt.Errorf("hostingService (hostingServiceName = %s) doesn't exist",
				hostingServiceName)
		}
		err := addApplicationRelatedHostingServiceToResult(&result1, applicationName, hostingServiceName, hostingService)
		if err != nil {
			return ProjectAnalysisResult{}, err
		}
	}
	// 3. Add application related backing Service
	for applicationName, backingServiceNames := range result2.ApplicationToBackingService {
		for backingServiceName := range backingServiceNames {
			backingService, ok := result2.Services[backingServiceName]
			if !ok {
				return ProjectAnalysisResult{}, fmt.Errorf("backingService (backingServiceName = %s) doesn't exist",
					backingServiceName)
			}
			err := addApplicationRelatedBackingServiceToResult(&result1, applicationName, backingServiceName, backingService)
			if err != nil {
				return ProjectAnalysisResult{}, err
			}
		}
	}
	return result1, nil
}

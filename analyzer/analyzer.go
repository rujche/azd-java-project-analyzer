package analyzer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func analyzeJavaProject(projectRootPath string) (ProjectAnalysisResult, error) {
	return analyzeJavaProjectSubDirectory(projectRootPath, projectRootPath)
}

func analyzeJavaProjectSubDirectory(projectRootPath string, subDirectoryPath string) (ProjectAnalysisResult, error) {
	entries, err := os.ReadDir(subDirectoryPath)
	if err != nil {
		return ProjectAnalysisResult{}, fmt.Errorf("reading directory: %w", err)
	}
	result := ProjectAnalysisResult{}
	for _, entry := range entries {
		if entry.IsDir() {
			newResult, err := analyzeJavaProjectSubDirectory(projectRootPath,
				filepath.Join(subDirectoryPath, entry.Name()))
			if err != nil {
				return ProjectAnalysisResult{}, fmt.Errorf("analyzing java project: %w", err)
			}
			result = mergeProject(result, newResult)
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
				result = mergeProject(result, newResult)
			}
		}
	}
	return result, nil
}

func analyzePomProject(projectRootPath string, pomFileAbsolutePath string) (ProjectAnalysisResult, error) {
	pom, err := createEffectivePom(pomFileAbsolutePath)
	if err != nil {
		return ProjectAnalysisResult{}, fmt.Errorf("creating effective pom: %w", err)
	}
	pomRelativePathPath, err := filepath.Rel(projectRootPath, pomFileAbsolutePath)
	if err != nil {
		return ProjectAnalysisResult{}, err
	}
	pom.pomFilePath = pomRelativePathPath
	if !isSpringBootRunnableProject(pom) {
		return ProjectAnalysisResult{}, nil
	}
	result := ProjectAnalysisResult{}
	projectPath := filepath.Dir(pomRelativePathPath)
	containerAppName := LabelName(filepath.Base(projectPath))
	result.resources = append(result.resources, Resource{containerAppName, AzureContainerApp})
	result.projectToResourceMappings = append(result.projectToResourceMappings,
		ProjectToResourceMapping{projectPath, containerAppName})
	for _, dep := range pom.Dependencies {
		if dep.GroupId == "com.mysql" && dep.ArtifactId == "mysql-connector-j" {
			// todo:
			// 1. support multiple container app use multiple mysql
			// 2. Support multiple container app use one mysql
			// 3. Same to other resources like postgresql
			mysqlResourceName := "mysql"
			result.resources = append(result.resources, Resource{mysqlResourceName, AzureDatabaseForMysql})
			result.resourceToResourceUsageBindings = append(result.resourceToResourceUsageBindings,
				ResourceToResourceUsageBinding{containerAppName, mysqlResourceName})
		}
		if dep.GroupId == "org.postgresql" && dep.ArtifactId == "postgresql" {
			postgresqlResourceName := "postgresql"
			result.resources = append(result.resources, Resource{postgresqlResourceName, AzureDatabaseForPostgresql})
			result.resourceToResourceUsageBindings = append(result.resourceToResourceUsageBindings,
				ResourceToResourceUsageBinding{containerAppName, postgresqlResourceName})
		}
		// todo: support other resource types.
	}
	return result, nil
}

func isSpringBootRunnableProject(pom pom) bool {
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

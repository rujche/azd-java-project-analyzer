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
	result := ProjectAnalysisResult{}
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
	applicationName := internal.LabelName(filepath.Base(projectRelativePath))
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
	for _, dep := range pom.Dependencies {
		if dep.GroupId == "com.mysql" && dep.ArtifactId == "mysql-connector-j" {
			// todo:
			// a. support multiple container app use multiple mysql
			// b. Support multiple container app use one mysql
			// c. Same to other resources like postgresql
			err = addApplicationRelatedBackingServiceToResult(&result, applicationName, DefaultMysqlServiceName,
				AzureDatabaseForMysql{})
			if err != nil {
				return result, err
			}
		}
		if dep.GroupId == "org.postgresql" && dep.ArtifactId == "postgresql" {
			err = addApplicationRelatedBackingServiceToResult(&result, applicationName, DefaultPostgresqlServiceName,
				AzureDatabaseForPostgresql{})
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

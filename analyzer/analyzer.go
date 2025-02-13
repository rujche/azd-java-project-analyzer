package analyzer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func analyzeJavaProject(projectRootPath string) (Project, error) {
	return analyzeJavaProjectSubDirectory(projectRootPath, projectRootPath)
}

func analyzeJavaProjectSubDirectory(projectRootPath string, subDirectoryPath string) (Project, error) {
	entries, err := os.ReadDir(subDirectoryPath)
	if err != nil {
		return Project{}, fmt.Errorf("reading directory: %w", err)
	}
	resultProject := Project{}
	for _, entry := range entries {
		if entry.IsDir() {
			newProject, err := analyzeJavaProjectSubDirectory(projectRootPath,
				filepath.Join(subDirectoryPath, entry.Name()))
			if err != nil {
				return Project{}, fmt.Errorf("analyzing java project: %w", err)
			}
			resultProject = mergeProject(resultProject, newProject)
		} else {
			// todo:
			// 1. Support file names like backend-pom.xml
			// 2. Support build.gradle
			if strings.ToLower(entry.Name()) == "pom.xml" {
				pomPath := filepath.Join(subDirectoryPath, entry.Name())
				newProject, err := analyzePomProject(projectRootPath, pomPath)
				if err != nil {
					return Project{}, err
				}
				// todo: consider multiple pom use same Azure resource
				resultProject = mergeProject(resultProject, newProject)
			}
		}
	}
	return resultProject, nil
}

func analyzePomProject(projectRootPath string, pomFileAbsolutePath string) (Project, error) {
	pom, err := createEffectivePom(pomFileAbsolutePath)
	if err != nil {
		return Project{}, fmt.Errorf("creating effective pom: %w", err)
	}
	pomRelativePathPath, err := filepath.Rel(projectRootPath, pomFileAbsolutePath)
	if err != nil {
		return Project{}, err
	}
	pom.pomFilePath = pomRelativePathPath
	if !isSpringBootRunnableProject(pom) {
		return Project{}, nil
	}
	project := Project{}
	projectPath := filepath.Dir(pomRelativePathPath)
	containerAppName := LabelName(filepath.Base(projectPath))
	project.resources = append(project.resources, Resource{containerAppName, AzureContainerApp})
	project.projectToResourceMappings = append(project.projectToResourceMappings,
		ProjectToResourceMapping{projectPath, containerAppName})
	for _, dep := range pom.Dependencies {
		if dep.GroupId == "com.mysql" && dep.ArtifactId == "mysql-connector-j" {
			// todo:
			// 1. support multiple container app use multiple mysql
			// 2. Support multiple container app use one mysql
			// 3. Same to other resources like postgresql
			mysqlResourceName := "mysql"
			project.resources = append(project.resources, Resource{mysqlResourceName, AzureDatabaseForMysql})
			project.resourceToResourceUsageBindings = append(project.resourceToResourceUsageBindings,
				ResourceToResourceUsageBinding{containerAppName, mysqlResourceName})
		}
		if dep.GroupId == "org.postgresql" && dep.ArtifactId == "postgresql" {
			postgresqlResourceName := "postgresql"
			project.resources = append(project.resources, Resource{postgresqlResourceName, AzureDatabaseForPostgresql})
			project.resourceToResourceUsageBindings = append(project.resourceToResourceUsageBindings,
				ResourceToResourceUsageBinding{containerAppName, postgresqlResourceName})
		}
		// todo: support other resource types.
	}
	return project, nil
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

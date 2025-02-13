package analyzer

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAnalyzeJavaProject(t *testing.T) {
	tests := []struct {
		name             string
		workingDirectory string
		expected         Project
	}{
		{
			name:             "java",
			workingDirectory: filepath.Join("testdata", "java"),
			expected:         Project{},
		},
		{
			name:             "java-multiple-modules",
			workingDirectory: filepath.Join("testdata", "java-multiple-modules"),
			expected: Project{
				resources: []Resource{
					{
						resourceName: "application",
						resourceType: AzureContainerApp,
					},
					{
						resourceName: "mysql",
						resourceType: AzureDatabaseForMysql,
					},
					{
						resourceName: "postgresql",
						resourceType: AzureDatabaseForPostgresql,
					},
				},
				resourceToResourceUsageBindings: []ResourceToResourceUsageBinding{
					{
						sourceResourceName: "application",
						targetResourceName: "mysql",
					},
					{
						sourceResourceName: "application",
						targetResourceName: "postgresql",
					},
				},
				projectToResourceMappings: []ProjectToResourceMapping{
					{
						projectRelativePath: "application",
						resourceName:        "application",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			project, err := analyzeJavaProject(tt.workingDirectory)
			if err != nil {
				t.Fatalf("analyzePomProject failed: %v", err)
			}

			require.Equal(t, tt.expected, project)
		})
	}
}

func TestAnalyzePomProject(t *testing.T) {
	tests := []struct {
		name     string
		testPoms []testPom
		expected Project
	}{
		{
			name: "not spring-boot runnable project",
			testPoms: []testPom{
				{
					pomFileRelativePath: "pom.xml",
					pomContentString: `
						<project>
							<modelVersion>4.0.0</modelVersion>
							<groupId>com.example</groupId>
							<artifactId>example-project</artifactId>
							<version>1.0.0</version>
							<dependencies>
								<dependency>
									<groupId>org.springframework</groupId>
									<artifactId>spring-core</artifactId>
									<version>5.3.8</version>
									<scope>compile</scope>
								</dependency>
								<dependency>
									<groupId>junit</groupId>
									<artifactId>junit</artifactId>
									<version>4.13.2</version>
									<scope>test</scope>
								</dependency>
							</dependencies>
						</project>
						`,
				},
			},
			expected: Project{},
		},
		{
			name: "has mysql and postgresql dependency",
			testPoms: []testPom{
				{
					pomFileRelativePath: filepath.Join("app-one", "pom.xml"),
					pomContentString: `
						<project>
							<modelVersion>4.0.0</modelVersion>
							<parent>
								<groupId>org.springframework.boot</groupId>
								<artifactId>spring-boot-starter-parent</artifactId>
								<version>3.3.0</version>
								<relativePath/> <!-- lookup parent from repository -->
							</parent>
							<groupId>com.example</groupId>
							<artifactId>example-project</artifactId>
							<version>1.0.0</version>
							<dependencies>
								<dependency>
									<groupId>com.mysql</groupId>
									<artifactId>mysql-connector-j</artifactId>
								</dependency>
								<dependency>
									<groupId>org.postgresql</groupId>
									<artifactId>postgresql</artifactId>
									<scope>test</scope>
								</dependency>
							</dependencies>
							<build>
								<plugins>
									<plugin>
										<groupId>org.springframework.boot</groupId>
										<artifactId>spring-boot-maven-plugin</artifactId>
									</plugin>
								</plugins>
							</build>
						</project>
						`,
				},
			},
			expected: Project{
				resources: []Resource{
					{
						resourceName: "app-one",
						resourceType: AzureContainerApp,
					},
					{
						resourceName: "mysql",
						resourceType: AzureDatabaseForMysql,
					},
					{
						resourceName: "postgresql",
						resourceType: AzureDatabaseForPostgresql,
					},
				},
				resourceToResourceUsageBindings: []ResourceToResourceUsageBinding{
					{
						sourceResourceName: "app-one",
						targetResourceName: "mysql",
					},
					{
						sourceResourceName: "app-one",
						targetResourceName: "postgresql",
					},
				},
				projectToResourceMappings: []ProjectToResourceMapping{
					{
						projectRelativePath: "app-one",
						resourceName:        "app-one",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			workingDir, err := prepareTestPomFiles(tt.testPoms)
			if err != nil {
				t.Fatalf("%v", err)
			}
			testPom := tt.testPoms[0]
			pomFileAbsolutePath := filepath.Join(workingDir, testPom.pomFileRelativePath)

			project, err := analyzePomProject(workingDir, pomFileAbsolutePath)
			if err != nil {
				t.Fatalf("analyzePomProject failed: %v", err)
			}

			require.Equal(t, tt.expected, project)
		})
	}
}

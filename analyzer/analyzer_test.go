package analyzer

import (
	"path/filepath"
	"testing"

	"ajpa/analyzer/internal"

	"github.com/stretchr/testify/require"
)

func TestAnalyzeJavaProject(t *testing.T) {
	tests := []struct {
		name             string
		workingDirectory string
		expected         ProjectAnalysisResult
	}{
		{
			name:             "java",
			workingDirectory: filepath.Join("testdata", "java"),
			expected: ProjectAnalysisResult{
				Name: "java",
			},
		},
		{
			name:             "java-multiple-modules",
			workingDirectory: filepath.Join("testdata", "java-multiple-modules"),
			expected: ProjectAnalysisResult{
				Name: "java-multiple-modules",
				Applications: map[string]Application{
					"application": {"application"},
				},
				Services: map[string]Service{
					"application":                AzureContainerApp{},
					DefaultMysqlServiceName:      AzureDatabaseForMysql{},
					DefaultPostgresqlServiceName: AzureDatabaseForPostgresql{},
				},
				ApplicationToHostingService: map[string]string{
					"application": "application",
				},
				ApplicationToBackingService: map[string]map[string]interface{}{
					"application": {
						DefaultMysqlServiceName:      "",
						DefaultPostgresqlServiceName: "",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			project, err := AnalyzeJavaProject(tt.workingDirectory)
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
		testPoms []internal.TestPom
		expected ProjectAnalysisResult
	}{
		{
			name: "not spring-boot runnable project",
			testPoms: []internal.TestPom{
				{
					PomFileRelativePath: "pom.xml",
					PomContentString: `
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
			expected: ProjectAnalysisResult{},
		},
		{
			name: "has mysql and postgresql dependency",
			testPoms: []internal.TestPom{
				{
					PomFileRelativePath: filepath.Join("application", "pom.xml"),
					PomContentString: `
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
			expected: ProjectAnalysisResult{
				Applications: map[string]Application{
					"application": {"application"},
				},
				Services: map[string]Service{
					"application":                AzureContainerApp{},
					DefaultMysqlServiceName:      AzureDatabaseForMysql{},
					DefaultPostgresqlServiceName: AzureDatabaseForPostgresql{},
				},
				ApplicationToHostingService: map[string]string{
					"application": "application",
				},
				ApplicationToBackingService: map[string]map[string]interface{}{
					"application": {
						DefaultMysqlServiceName:      "",
						DefaultPostgresqlServiceName: "",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			workingDir, err := internal.PrepareTestPomFiles(tt.testPoms)
			if err != nil {
				t.Fatalf("%v", err)
			}
			testPom := tt.testPoms[0]
			pomFileAbsolutePath := filepath.Join(workingDir, testPom.PomFileRelativePath)

			project, err := analyzePomProject(workingDir, pomFileAbsolutePath)
			if err != nil {
				t.Fatalf("analyzePomProject failed: %v", err)
			}

			require.Equal(t, tt.expected, project)
		})
	}
}

package converter

import (
	"testing"

	"ajpa/analyzer"

	"github.com/azure/azure-dev/cli/azd/pkg/project"
	"github.com/stretchr/testify/require"
)

func TestProjectAnalysisResultToAzdProjectConfig(t *testing.T) {
	tests := []struct {
		name     string
		result   analyzer.ProjectAnalysisResult
		expected project.ProjectConfig
	}{
		{
			name: "one dependency: postgresql",
			result: analyzer.ProjectAnalysisResult{
				Name: "app-one-sample",
				Applications: map[string]analyzer.Application{
					"app-one": {ProjectRelativePath: "app-one"},
				},
				Services: map[string]analyzer.Service{
					"app-one":                             analyzer.AzureContainerApp{},
					analyzer.DefaultPostgresqlServiceName: analyzer.AzureDatabaseForPostgresql{},
				},
				ApplicationToHostingService: map[string]string{
					"app-one": "app-one",
				},
				ApplicationToBackingService: map[string]map[string]interface{}{
					"app-one": {
						analyzer.DefaultPostgresqlServiceName: "",
					},
				},
			},
			expected: project.ProjectConfig{
				Name: "app-one-sample",
				Services: map[string]*project.ServiceConfig{
					"app-one": {
						Project:      nil, // will be updated in test
						Name:         "app-one",
						Language:     project.ServiceLanguageJava,
						RelativePath: "app-one",
						Host:         project.ContainerAppTarget,
					},
				},
				Resources: map[string]*project.ResourceConfig{
					"app-one": {
						Project: nil, // will be updated in test
						Type:    project.ResourceTypeHostContainerApp,
						Name:    "app-one",
						Uses:    []string{analyzer.DefaultPostgresqlServiceName},
						Props: project.ContainerAppProps{
							Port: 8080,
						},
					},
					analyzer.DefaultPostgresqlServiceName: {
						Project: nil, // will be updated in test
						Type:    project.ResourceTypeDbPostgres,
						Name:    analyzer.DefaultPostgresqlServiceName,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key := range tt.expected.Services {
				tt.expected.Services[key].Project = &tt.expected
			}
			for key := range tt.expected.Resources {
				tt.expected.Resources[key].Project = &tt.expected
			}
			config, err := ProjectAnalysisResultToAzdProjectConfig(tt.result)
			if err != nil {
				t.Fatalf("ProjectAnalysisResultToAzdProjectConfig failed: %v", err)
			}
			require.Equal(t, tt.expected, config)
		})
	}
}

package converter

import (
	"testing"

	"ajpa/analyzer"
	"ajpa/converter/azd"

	"github.com/stretchr/testify/require"
)

func TestProjectAnalysisResultToAzdProjectConfig(t *testing.T) {
	tests := []struct {
		name     string
		result   analyzer.ProjectAnalysisResult
		expected azd.ProjectConfig
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
			expected: azd.ProjectConfig{
				Name: "app-one-sample",
				Services: map[string]*azd.ServiceConfig{
					"app-one": {
						Name:         "app-one",
						Language:     azd.ServiceLanguageJava,
						RelativePath: "app-one",
						Host:         azd.ContainerAppTarget,
					},
				},
				Resources: map[string]*azd.ResourceConfig{
					"app-one": {
						Type: azd.ResourceTypeHostContainerApp,
						Name: "app-one",
						Uses: []string{analyzer.DefaultPostgresqlServiceName},
						Props: azd.ContainerAppProps{
							Port: 8080,
						},
					},
					analyzer.DefaultPostgresqlServiceName: {
						Type: azd.ResourceTypeDbPostgres,
						Name: analyzer.DefaultPostgresqlServiceName,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := ProjectAnalysisResultToAzdProjectConfig(tt.result)
			if err != nil {
				t.Fatalf("ProjectAnalysisResultToAzdProjectConfig failed: %v", err)
			}
			require.Equal(t, tt.expected, config)
		})
	}
}

package internal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadProperties(t *testing.T) {
	var properties = ReadProperties(filepath.Join("testdata", "java-spring", "project-one"))
	require.Equal(t, "", properties["not.exist"])
	require.Equal(t, "jdbc:h2:mem:testdb", properties["spring.datasource.url"])

	properties = ReadProperties(filepath.Join("testdata", "java-spring", "project-two"))
	require.Equal(t, "", properties["not.exist"])
	require.Equal(t, "jdbc:h2:mem:testdb", properties["spring.datasource.url"])

	properties = ReadProperties(filepath.Join("testdata", "java-spring", "project-three"))
	require.Equal(t, "", properties["not.exist"])
	require.Equal(t, "HTML", properties["spring.thymeleaf.mode"])

	properties = ReadProperties(filepath.Join("testdata", "java-spring", "project-four"))
	require.Equal(t, "", properties["not.exist"])
	require.Equal(t, "mysql", properties["database"])
}

func TestGetEnvironmentVariablePlaceholderHandledValue(t *testing.T) {
	tests := []struct {
		name                 string
		inputValue           string
		environmentVariables map[string]string
		expectedValue        string
	}{
		{
			"No environment variable placeholder",
			"valueOne",
			map[string]string{},
			"valueOne",
		},
		{
			"Has invalid environment variable placeholder",
			"${VALUE_ONE",
			map[string]string{},
			"${VALUE_ONE",
		},
		{
			"Has valid environment variable placeholder, but environment variable not set",
			"${VALUE_TWO}",
			map[string]string{},
			"",
		},
		{
			"Has valid environment variable placeholder, and environment variable set",
			"${VALUE_THREE}",
			map[string]string{"VALUE_THREE": "valueThree"},
			"valueThree",
		},
		{
			"Has valid environment variable placeholder with default value, but environment variable not set",
			"${VALUE_TWO:defaultValue}",
			map[string]string{},
			"defaultValue",
		},
		{
			"Has valid environment variable placeholder with default value, and environment variable set",
			"${VALUE_THREE:defaultValue}",
			map[string]string{"VALUE_THREE": "valueThree"},
			"valueThree",
		},
		{
			"Has multiple environment variable placeholder with default value, and environment variable not set",
			"jdbc:mysql://${MYSQL_HOST:localhost}:${MYSQL_PORT:3306}/${MYSQL_DATABASE:pet-clinic}",
			map[string]string{},
			"jdbc:mysql://localhost:3306/pet-clinic",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.environmentVariables {
				err := os.Setenv(k, v)
				require.NoError(t, err)
			}
			handledValue := getEnvironmentVariablePlaceholderHandledValue(tt.inputValue)
			require.Equal(t, tt.expectedValue, handledValue)
		})
	}
}

func TestGetDatabaseName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"jdbc:postgresql://localhost:5432/your-database-name", "your-database-name"},
		{"jdbc:postgresql://remote_host:5432/your-database-name", "your-database-name"},
		{"jdbc:postgresql://your_postgresql_server:5432/your-database-name?sslmode=require", "your-database-name"},
		{
			"jdbc:postgresql://your_postgresql_server.postgres.database.azure.com:5432/your-database-name?sslmode=require",
			"your-database-name",
		},
		{
			"jdbc:postgresql://your_postgresql_server:5432/your-database-name?user=your_username&password=your_password",
			"your-database-name",
		},
		{
			"jdbc:postgresql://your_postgresql_server.postgres.database.azure.com:5432/your-database-name" +
				"?sslmode=require&spring.datasource.azure.passwordless-enabled=true", "your-database-name",
		},
	}
	for _, test := range tests {
		result := GetDatabaseName(test.input)
		if result != test.expected {
			t.Errorf("For input '%s', expected '%s', but got '%s'", test.input, test.expected, result)
		}
	}
}

func TestIsValidDatabaseName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"InvalidNameWithUnderscore", "invalid_name", false},
		{"TooShortName", "sh", false},
		{
			"TooLongName", "this-name-is-way-too-long-to-be-considered-valid-" +
				"because-it-exceeds-sixty-three-characters", false,
		},
		{"InvalidStartWithHyphen", "-invalid-start", false},
		{"InvalidEndWithHyphen", "invalid-end-", false},
		{"ValidName", "valid-name", true},
		{"ValidNameWithNumbers", "valid123-name", true},
		{"ValidNameWithOnlyLetters", "valid-name", true},
		{"ValidNameWithOnlyNumbers", "123456", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := IsValidDatabaseName(test.input)
			if result != test.expected {
				t.Errorf("For input '%s', expected %v, but got %v", test.input, test.expected, result)
			}
		})
	}
}

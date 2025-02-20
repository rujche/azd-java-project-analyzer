package azd

import (
	"bytes"
	"fmt"
	"os"

	"github.com/braydonk/yaml"
)

type ProjectConfig struct {
	Name      string                     `yaml:"name"`
	Services  map[string]*ServiceConfig  `yaml:"services,omitempty"`
	Resources map[string]*ResourceConfig `yaml:"resources,omitempty"`
}

type ServiceTargetKind string

const (
	NonSpecifiedTarget       ServiceTargetKind = ""
	AppServiceTarget         ServiceTargetKind = "appservice"
	ContainerAppTarget       ServiceTargetKind = "containerapp"
	AzureFunctionTarget      ServiceTargetKind = "function"
	StaticWebAppTarget       ServiceTargetKind = "staticwebapp"
	SpringAppTarget          ServiceTargetKind = "springapp"
	AksTarget                ServiceTargetKind = "aks"
	DotNetContainerAppTarget ServiceTargetKind = "containerapp-dotnet"
	AiEndpointTarget         ServiceTargetKind = "ai.endpoint"
)

type ServiceLanguageKind string

const (
	ServiceLanguageNone       ServiceLanguageKind = ""
	ServiceLanguageDotNet     ServiceLanguageKind = "dotnet"
	ServiceLanguageCsharp     ServiceLanguageKind = "csharp"
	ServiceLanguageFsharp     ServiceLanguageKind = "fsharp"
	ServiceLanguageJavaScript ServiceLanguageKind = "js"
	ServiceLanguageTypeScript ServiceLanguageKind = "ts"
	ServiceLanguagePython     ServiceLanguageKind = "python"
	ServiceLanguageJava       ServiceLanguageKind = "java"
	ServiceLanguageDocker     ServiceLanguageKind = "docker"
	ServiceLanguageSwa        ServiceLanguageKind = "swa"
)

type ServiceConfig struct {
	Name         string              `yaml:"-"`
	RelativePath string              `yaml:"project"`
	Host         ServiceTargetKind   `yaml:"host"`
	Language     ServiceLanguageKind `yaml:"language"`
}

type ResourceType string

const (
	ResourceTypeDbRedis             ResourceType = "db.redis"
	ResourceTypeDbPostgres          ResourceType = "db.postgres"
	ResourceTypeDbMongo             ResourceType = "db.mongo"
	ResourceTypeHostContainerApp    ResourceType = "host.containerapp"
	ResourceTypeOpenAiModel         ResourceType = "ai.openai.model"
	ResourceTypeMessagingEventHubs  ResourceType = "messaging.eventhubs"
	ResourceTypeMessagingServiceBus ResourceType = "messaging.servicebus"
	ResourceTypeStorage             ResourceType = "storage"
)

type ResourceConfig struct {
	// Type of resource
	Type ResourceType `yaml:"type"`
	// The name of the resource
	Name string `yaml:"-"`
	// The properties for the resource
	RawProps map[string]yaml.Node `yaml:",inline"`
	Props    interface{}          `yaml:"-"`
	// Relationships to other resources
	Uses []string `yaml:"uses,omitempty"`
}

func (r *ResourceConfig) MarshalYAML() (interface{}, error) {
	type rawResourceConfig ResourceConfig
	raw := rawResourceConfig(*r)

	var marshalRawProps = func(in interface{}) error {
		marshaled, err := yaml.Marshal(in)
		if err != nil {
			return fmt.Errorf("marshaling props: %w", err)
		}

		props := map[string]yaml.Node{}
		if err := yaml.Unmarshal(marshaled, &props); err != nil {
			return err
		}
		raw.RawProps = props
		return nil
	}

	var errMarshal error
	switch raw.Type {
	case ResourceTypeOpenAiModel:
		errMarshal = marshalRawProps(raw.Props.(AIModelProps))
	case ResourceTypeHostContainerApp:
		errMarshal = marshalRawProps(raw.Props.(ContainerAppProps))
	case ResourceTypeMessagingEventHubs:
		errMarshal = marshalRawProps(raw.Props.(EventHubsProps))
	case ResourceTypeMessagingServiceBus:
		errMarshal = marshalRawProps(raw.Props.(ServiceBusProps))
	case ResourceTypeStorage:
		errMarshal = marshalRawProps(raw.Props.(StorageProps))
	}

	if errMarshal != nil {
		return nil, errMarshal
	}

	return raw, nil
}

func (r *ResourceConfig) UnmarshalYAML(value *yaml.Node) error {
	type rawResourceConfig ResourceConfig
	raw := rawResourceConfig{}
	if err := value.Decode(&raw); err != nil {
		return err
	}

	var unmarshalProps = func(v interface{}) error {
		value, err := yaml.Marshal(raw.RawProps)
		if err != nil {
			return fmt.Errorf("failed to marshal raw props: %w", err)
		}

		if err := yaml.Unmarshal(value, v); err != nil {
			return err
		}

		return nil
	}

	// Unmarshal props based on type
	switch raw.Type {
	case ResourceTypeOpenAiModel:
		amp := AIModelProps{}
		if err := unmarshalProps(&amp); err != nil {
			return err
		}
		raw.Props = amp
	case ResourceTypeHostContainerApp:
		cap := ContainerAppProps{}
		if err := unmarshalProps(&cap); err != nil {
			return err
		}
		raw.Props = cap
	case ResourceTypeMessagingEventHubs:
		ehp := EventHubsProps{}
		if err := unmarshalProps(&ehp); err != nil {
			return err
		}
		raw.Props = ehp
	case ResourceTypeMessagingServiceBus:
		sbp := ServiceBusProps{}
		if err := unmarshalProps(&sbp); err != nil {
			return err
		}
		raw.Props = sbp
	case ResourceTypeStorage:
		sp := StorageProps{}
		if err := unmarshalProps(&sp); err != nil {
			return err
		}
		raw.Props = sp
	}

	*r = ResourceConfig(raw)
	return nil
}

type ContainerAppProps struct {
	Port int             `yaml:"port,omitempty"`
	Env  []ServiceEnvVar `yaml:"env,omitempty"`
}

type ServiceEnvVar struct {
	Name string `yaml:"name,omitempty"`

	// either Value or Secret can be set, but not both
	Value  string `yaml:"value,omitempty"`
	Secret string `yaml:"secret,omitempty"`
}

type AIModelProps struct {
	Model AIModelPropsModel `yaml:"model,omitempty"`
}

type AIModelPropsModel struct {
	Name    string `yaml:"name,omitempty"`
	Version string `yaml:"version,omitempty"`
}

type ServiceBusProps struct {
	Queues []string `yaml:"queues,omitempty"`
	Topics []string `yaml:"topics,omitempty"`
}

type EventHubsProps struct {
	Hubs []string `yaml:"hubs,omitempty"`
}

type StorageProps struct {
	Containers []string `yaml:"containers,omitempty"`
}

func Save(projectConfig *ProjectConfig, projectFilePath string) error {
	projectBytes, err := yaml.Marshal(projectConfig)
	if err != nil {
		return fmt.Errorf("marshalling project yaml: %w", err)
	}
	version := "v1.0"
	annotation := fmt.Sprintf(
		"# yaml-language-server: $schema=https://raw.githubusercontent.com/Azure/azure-dev/main/schemas/%s/azure.yaml.json",
		version)
	projectFileContents := bytes.NewBufferString(annotation + "\n\n")
	_, err = projectFileContents.Write(projectBytes)
	if err != nil {
		return fmt.Errorf("preparing new project file contents: %w", err)
	}
	err = os.WriteFile(projectFilePath, projectFileContents.Bytes(), 0644)
	if err != nil {
		return fmt.Errorf("saving project file: %w", err)
	}

	return nil
}

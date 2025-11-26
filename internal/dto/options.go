package dto

// GenerationOptions — внешние параметры генерации (CI/Registry/Sonar/Nexus).
type GenerationOptions struct {
	CI               string `json:"ci"` // "gitlab"|"jenkins"
	RegistryURL      string `json:"registry_url"`
	RegistryProject  string `json:"registry_project"`
	RegistryUser     string `json:"registry_user"`
	RegistryPassword string `json:"registry_password"`

	SonarHost  string `json:"sonar_host"`
	SonarToken string `json:"sonar_token"`

	NexusURL      string `json:"nexus_url"`
	NexusUser     string `json:"nexus_user"`
	NexusPassword string `json:"nexus_password"`
}

package dto

// DockerMeta — предпочтения/детекты контейнеризации.
type DockerMeta struct {
	DockerfileDetected bool   `json:"dockerfile_detected"`
	ComposeDetected    bool   `json:"compose_detected"`
	PreferredBase      string `json:"preferred_base"` // "alpine"|"slim"|"distroless"|"scratch"
	ExposedPort        int    `json:"exposed_port"`
	ImageName          string `json:"image_name"`
}

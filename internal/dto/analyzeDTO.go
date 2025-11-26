package dto

type AnalyzeDTO struct {
	IsGoModule bool
	// IsDocker   bool
	IsNodeJS  bool
	IsPython   bool
	IsJava     bool
}

type GoDTO struct {
	Version string
	IsDocker bool
	IsDockerCompose bool
}
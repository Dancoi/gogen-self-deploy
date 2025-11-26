package dto

// RepoDTO описывает источник и локальные пути к репозиторию.
type RepoDTO struct {
	RepoURL   string `json:"repo_url"`
	OutputDir string `json:"output_dir"` // базовая папка для вывода
	LocalPath string `json:"local_path"` // путь, куда клонировали
	RepoName  string `json:"repo_name"`  // опционально
}

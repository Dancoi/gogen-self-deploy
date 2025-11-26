package analyzer

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-enry/go-enry/v2"
)

func AnalyzeGlobalStats(result *ProjectAnalysisResult, root string) {
	result.Languages = make(map[string]float64)
	result.Infrastructure = make([]string, 0)

	langStats := make(map[string]int64)
	var totalBytes int64
	infraMap := make(map[string]bool)

	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if d.IsDir() {
			if shouldSkipDir(d.Name()) || enry.IsVendor(path) {
				return filepath.SkipDir
			}
			return nil
		}

		// 1. Инфраструктура
		name := strings.ToLower(d.Name())
		rel, _ := filepath.Rel(root, path)

		if name == "dockerfile" || strings.HasPrefix(name, "docker-compose") {
			infraMap["Docker"] = true
		} else if name == "kustomization.yaml" || strings.HasSuffix(name, "chart.yaml") {
			infraMap["Kubernetes"] = true
		} else if strings.HasPrefix(rel, ".github/workflows") {
			infraMap["GitHub Actions"] = true
		} else if name == ".gitlab-ci.yml" {
			infraMap["GitLab CI"] = true
		} else if name == "jenkinsfile" {
			infraMap["Jenkins"] = true
		}

		// 2. Языки
		if !enry.IsVendor(path) && !enry.IsGenerated(path, nil) {
			info, _ := d.Info()
			if info.Size() > 0 {
				lang, _ := enry.GetLanguageByExtension(d.Name())
				if lang == "" {
					content, _ := os.ReadFile(path)
					lang = enry.GetLanguage(d.Name(), content)
				}

				// ИСПРАВЛЕНИЕ: Функция возвращает только одно значение
				langType := enry.GetLanguageType(lang)

				// Фильтр: Считаем только Programming и Markup (HTML/CSS)
				if langType == enry.Programming || lang == "HTML" || lang == "CSS" || lang == "SCSS" {
					langStats[lang] += info.Size()
					totalBytes += info.Size()
				}
			}
		}

		return nil
	})

	// Подсчет процентов
	for lang, size := range langStats {
		percent := (float64(size) / float64(totalBytes)) * 100
		if percent > 1.0 { // Отсекаем шум < 1%
			result.Languages[lang] = percent
		}
	}

	for k := range infraMap {
		result.Infrastructure = append(result.Infrastructure, k)
	}
}

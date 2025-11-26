package analyzer

import (
	"github.com/Dancoi/gogen-self-deploy/internal/dto"
)

func AnalyzRepo(dto dto.RepoDTO) (result *ProjectAnalysisResult, err error) {
	root := dto.OutputDir
	if dto.RepoName != "" {
		root = root + "/" + dto.RepoName
	}

	// Инициализация
	result = &ProjectAnalysisResult{
		RepositoryName: dto.RepoName,
		Modules:        []*ProjectModule{},
		Languages:      make(map[string]float64),
		Infrastructure: []string{},
	}

	// 1. Глобальный анализ (Языки + Инфраструктура)
	AnalyzeGlobalStats(result, root)

	// 2. Анализ модулей
	AnalyzeGoModule(result, root)
	AnalyzeJavaModule(result, root) // Внутри теперь есть фильтр от "мусорных" модулей
	AnalyzeNodeModule(result, root)
	AnalyzePythonModule(result, root)

	// 3. Стратегия
	if len(result.Modules) > 1 {
		result.PipelineStrategy = PipelineStrategyMonorepo
	} else {
		result.PipelineStrategy = PipelineStrategyStandalone
	}

	// Попытка определить главный фреймворк
	if len(result.Modules) > 0 {
		for _, m := range result.Modules {
			// Приоритет модулю, у которого определен фреймворк
			if m.Framework != "" {
				result.MainFramework = m.Framework
				result.MainFrameworkVersion = m.FrameworkVersion
				break
			}
		}
		// Если ничего не нашли, берем фреймворк первого модуля (даже если пусто)
		if result.MainFramework == "" {
			result.MainFramework = result.Modules[0].Framework
			result.MainFrameworkVersion = result.Modules[0].FrameworkVersion
		}
	}

	return result, nil
}

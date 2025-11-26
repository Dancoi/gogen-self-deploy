package analyzer

import "github.com/Dancoi/gogen-self-deploy/internal/dto"

// func AnalyzRepo(result *ProjectAnalysisResult) {
// 	AnalyzeGoModule(result)
// 	// AnalyzeJavaModule(result)
// 	// AnalyzeNodeModule(result)
// 	// AnalyzePythonModule(result)

// 	if len(result.Modules) > 1 {
// 		result.PipelineStrategy = PipelineStrategyMonorepo
// 	} else {
// 		result.PipelineStrategy = PipelineStrategyStandalone
// 		if len(result.Modules) == 1 {
// 			result.MainFramework = result.Modules[0].Framework
// 			result.MainFrameworkVersion = result.Modules[0].FrameworkVersion
// 		}
// 	}
// }

func AnalyzRepo(dto dto.RepoDTO) (result *ProjectAnalysisResult, err error) {
	// Initialize result to avoid nil-pointer dereference in analyzers
	result = &ProjectAnalysisResult{
		Modules: []*ProjectModule{},
	}

	AnalyzeGoModule(dto, result)
	// AnalyzeJavaModule(result)
	// AnalyzeNodeModule(result)
	// AnalyzePythonModule(result)

	if len(result.Modules) > 1 {
		result.PipelineStrategy = PipelineStrategyMonorepo
	} else {
		result.PipelineStrategy = PipelineStrategyStandalone
		if len(result.Modules) == 1 {
			result.MainFramework = result.Modules[0].Framework
			result.MainFrameworkVersion = result.Modules[0].FrameworkVersion
		}
	}

	return result, nil
}

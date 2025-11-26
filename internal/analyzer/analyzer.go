package analyzer

func AnalyzRepo(result *ProjectAnalysisResult) {
	AnalyzeGoModule(result)
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
}

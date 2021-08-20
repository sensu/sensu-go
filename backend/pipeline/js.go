package pipeline

const pipelineRoleName = "system:pipeline"

// PipelineFilterFuncs gets patched by enterprise sensu-go
var PipelineFilterFuncs map[string]interface{}

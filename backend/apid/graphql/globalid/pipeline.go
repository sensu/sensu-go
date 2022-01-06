package globalid

import corev2 "github.com/sensu/sensu-go/api/core/v2"

var pipelineName = "corev2/pipeline"

// PipelineTranslator global ID resource
var PipelineTranslator = commonTranslator{
	name:       pipelineName,
	encodeFunc: standardEncoder(pipelineName, "Name"),
	decodeFunc: standardDecoder,
	isResponsibleFunc: func(record interface{}) bool {
		_, ok := record.(*corev2.Pipeline)
		return ok
	},
}

// Register entity encoder/decoder
func init() { RegisterTranslator(PipelineTranslator) }

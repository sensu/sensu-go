package sensu

import "github.com/sensu/sensu-go/api/core/v2"



#Wrapper: {
	type:        string
	api_version: "core/v2"
	metadata:    v2.#ObjectMeta
	spec:        {...}
}

_types: [v2.#APIKey,v2.#Asset,v2.#CheckConfig,v2.#Entity,v2.#Handler,v2.#Mutator,v2.#Namespace,v2.#Pipeline,v2.#Role,v2.#Secret]

#TypeWrapper: #APIKeyWrapper | #AssetWrapper | #CheckConfigWrapper | #EntityWrapper | #HandlerWrapper | #MutatorWrapper | #NamespaceWrapper | #PipelineWrapper | #RoleWrapper | #SecretWrapper

#APIKeyWrapper: #Wrapper & {
    type: "APIKey"
    spec: v2.#APIKey
}
#AssetWrapper: #Wrapper & {
    type: "Asset"
    spec: v2.#Asset
}
#CheckConfigWrapper: #Wrapper & {
    type: "CheckConfig"
    spec: v2.#CheckConfig
}
#EntityWrapper: #Wrapper & {
    type: "Entity"
    spec: v2.#Entity
}
#HandlerWrapper: #Wrapper & {
    type: "Handler"
    spec: v2.#Handler
}
#MutatorWrapper: #Wrapper & {
    type: "Mutator"
    spec: v2.#Mutator
}
#NamespaceWrapper: #Wrapper & {
    type: "Namespace"
    spec: v2.#Namespace
}
#PipelineWrapper: #Wrapper & {
    type: "Pipeline"
    spec: v2.#Pipeline
}
#RoleWrapper: #Wrapper & {
    type: "Role"
    spec: v2.#Role
}
#SecretWrapper: #Wrapper & {
    type: "Secret"
    spec: v2.#Secret
}


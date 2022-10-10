package graphql

import (
	"context"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/apid/graphql/relay"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/sensu-go/graphql/tracing"
)

type InitHook func(*graphql.Service, ServiceConfig)

// InitHooks allow consumers to hook into the initialization of the service and
// mutate the schema. Useful for product variants.
var InitHooks = []InitHook{}

// ClientFactory instantiates new instances of the REST API client
type ClientFactory interface {
	NewWithContext(ctx context.Context) client.APIClient
}

// ServiceConfig describes values required to instantiate service.
type ServiceConfig struct {
	AssetClient        AssetClient
	CheckClient        CheckClient
	EntityClient       EntityClient
	EventClient        EventClient
	EventFilterClient  EventFilterClient
	GetBackendEntity   func() (interface{ GetMetadata() *corev2.ObjectMeta }, error)
	HandlerClient      HandlerClient
	HealthController   EtcdHealthController
	MutatorClient      MutatorClient
	SilencedClient     SilencedClient
	NamespaceClient    NamespaceClient
	HookClient         HookClient
	UserClient         UserClient
	RBACClient         RBACClient
	VersionController  VersionController
	GenericClient      GenericClient
	MetricGatherer     MetricGatherer
	ClusterMetricStore ClusterMetricStore
}

// Service describes the Sensu GraphQL service capable of handling queries.
type Service struct {
	Target       *graphql.Service
	Config       *ServiceConfig
	NodeRegister *relay.NodeRegister
}

// NewService instantiates new GraphQL service
func NewService(cfg ServiceConfig) (*Service, error) {
	svc := graphql.NewService()

	nodeRegister := relay.NodeRegister{}
	nodeResolver := relay.Resolver{Register: &nodeRegister}
	wrapper := Service{
		Target:       svc,
		Config:       &cfg,
		NodeRegister: &nodeRegister,
	}

	// Register types w/ generic node resolver
	registerNodeResolvers(nodeRegister, cfg)

	// Register types
	schema.RegisterAsset(svc, &assetImpl{})
	schema.RegisterBackendEntity(svc, &backendEntityImpl{})
	schema.RegisterCoreV2AssetBuild(svc, &schema.CoreV2AssetBuildAliases{})
	schema.RegisterCoreV2Deregistration(svc, &schema.CoreV2DeregistrationAliases{})
	schema.RegisterCoreV2Network(svc, &schema.CoreV2NetworkAliases{})
	schema.RegisterCoreV2NetworkInterface(svc, &schema.CoreV2NetworkInterfaceAliases{})
	schema.RegisterCoreV2Pipeline(svc, &corev2PipelineImpl{})
	schema.RegisterCoreV2PipelineExtensionOverrides(svc, &corev2PipelineImpl{})
	schema.RegisterCoreV2PipelineWorkflow(svc, &schema.CoreV2PipelineWorkflowAliases{})
	schema.RegisterCoreV2Process(svc, &schema.CoreV2ProcessAliases{})
	schema.RegisterCoreV2ResourceReference(svc, &schema.CoreV2ResourceReferenceAliases{})
	schema.RegisterCoreV2Secret(svc, &schema.CoreV2SecretAliases{})
	schema.RegisterCoreV2System(svc, &schema.CoreV2SystemAliases{})
	schema.RegisterCoreV3EntityConfig(svc, &corev3EntityConfigImpl{})
	schema.RegisterCoreV3EntityConfigExtensionOverrides(svc, &corev3EntityConfigExtImpl{client: cfg.GenericClient, entityClient: cfg.EntityClient})
	schema.RegisterCoreV3EntityState(svc, &corev3EntityStateImpl{})
	schema.RegisterCoreV3EntityStateExtensionOverrides(svc, &corev3EntityStateExtImpl{client: cfg.GenericClient, entityClient: cfg.EntityClient})
	schema.RegisterNamespace(svc, &namespaceImpl{client: cfg.NamespaceClient, entityClient: cfg.EntityClient, eventClient: cfg.EventClient, serviceConfig: &cfg})
	schema.RegisterErrCode(svc)
	schema.RegisterEvent(svc, &eventImpl{})
	schema.RegisterEventsListOrder(svc)
	schema.RegisterJSON(svc, jsonImpl{})
	schema.RegisterKVPairString(svc, &schema.KVPairStringAliases{})
	schema.RegisterQuery(svc, &queryImpl{nodeResolver: &nodeResolver, svc: cfg})
	schema.RegisterMutator(svc, &mutatorImpl{})
	schema.RegisterMutatorConnection(svc, &schema.MutatorConnectionAliases{})
	schema.RegisterMutatorListOrder(svc)
	schema.RegisterNode(svc, &nodeImpl{&nodeResolver})
	schema.RegisterNamespaced(svc, nil)
	schema.RegisterObjectMeta(svc, &objectMetaImpl{})
	schema.RegisterOffsetPageInfo(svc, &offsetPageInfoImpl{})
	schema.RegisterProxyRequests(svc, &schema.ProxyRequestsAliases{})
	schema.RegisterResource(svc, nil)
	schema.RegisterResolveEventPayload(svc, &schema.ResolveEventPayloadAliases{})
	schema.RegisterSchema(svc)
	schema.RegisterSilenceable(svc, nil)
	schema.RegisterSilenced(svc, &silencedImpl{client: cfg.CheckClient})
	schema.RegisterSilencedConnection(svc, &schema.SilencedConnectionAliases{})
	schema.RegisterSilencesListOrder(svc)
	schema.RegisterSuggestionOrder(svc)
	schema.RegisterSuggestionResultSet(svc, &schema.SuggestionResultSetAliases{})
	schema.RegisterUint(svc, unsignedIntegerImpl{})
	schema.RegisterViewer(svc, &viewerImpl{userClient: cfg.UserClient})

	// Register check types
	schema.RegisterCheck(svc, &checkImpl{})
	schema.RegisterCheckConfig(svc, &checkCfgImpl{})
	schema.RegisterCheckConfigConnection(svc, &schema.CheckConfigConnectionAliases{})
	schema.RegisterCheckHistory(svc, &checkHistoryImpl{})
	schema.RegisterCheckListOrder(svc)

	// Register entity types
	schema.RegisterEntity(svc, &entityImpl{})
	schema.RegisterEntityConnection(svc, &schema.EntityConnectionAliases{})
	schema.RegisterEntityListOrder(svc)
	schema.RegisterDeregistration(svc, &deregistrationImpl{})
	schema.RegisterNetwork(svc, &networkImpl{})
	schema.RegisterNetworkInterface(svc, &networkInterfaceImpl{})
	schema.RegisterProcess(svc, &processImpl{})
	schema.RegisterSystem(svc, &systemImpl{})

	// Register event types
	schema.RegisterEvent(svc, &eventImpl{})
	schema.RegisterEventConnection(svc, &schema.EventConnectionAliases{})

	// Register event filter types
	schema.RegisterEventFilter(svc, &eventFilterImpl{})
	schema.RegisterEventFilterConnection(svc, &schema.EventFilterConnectionAliases{})
	schema.RegisterEventFilterAction(svc)
	schema.RegisterEventFilterListOrder(svc)

	// Register hook types
	schema.RegisterHook(svc, &hookImpl{})
	schema.RegisterHookConfig(svc, &hookCfgImpl{})
	schema.RegisterHookList(svc, &hookListImpl{})

	// Register handler types
	schema.RegisterHandler(svc, &handlerImpl{client: cfg.MutatorClient})
	schema.RegisterHandlerListOrder(svc)
	schema.RegisterHandlerConnection(svc, &schema.HandlerConnectionAliases{})
	schema.RegisterHandlerSocket(svc, &handlerSocketImpl{})

	// Register health types
	schema.RegisterClusterHealth(svc, &clusterHealthImpl{healthController: cfg.HealthController})
	schema.RegisterEtcdAlarmMember(svc, &etcdAlarmMemberImpl{})
	schema.RegisterEtcdAlarmType(svc)
	schema.RegisterEtcdClusterHealth(svc, &etcdClusterHealthImpl{})
	schema.RegisterEtcdClusterMemberHealth(svc, &etcdClusterMemberHealthImpl{})

	// Register metrics
	schema.RegisterBucketMetric(svc, &schema.BucketMetricAliases{})
	schema.RegisterCounterMetric(svc, &counterMetricImpl{})
	schema.RegisterGaugeMetric(svc, &gaugeMetricImpl{})
	schema.RegisterHistogramMetric(svc, &histogramMetricImpl{})
	schema.RegisterMetric(svc, &metricImpl{})
	schema.RegisterMetricFamily(svc, &metricFamilyImpl{})
	schema.RegisterMetricKind(svc)
	schema.RegisterQuantileMetric(svc, &schema.QuantileMetricAliases{})
	schema.RegisterSummaryMetric(svc, &summaryMetricImpl{})
	schema.RegisterUntypedMetric(svc, &untypedMetricImpl{})

	// Register time window
	schema.RegisterTimeWindowDays(svc, &schema.TimeWindowDaysAliases{})
	schema.RegisterTimeWindowWhen(svc, &schema.TimeWindowWhenAliases{})
	schema.RegisterTimeWindowTimeRange(svc, &schema.TimeWindowTimeRangeAliases{})

	// Register RBAC types
	schema.RegisterCoreV2ClusterRole(svc, &clusterRoleImpl{})
	schema.RegisterCoreV2ClusterRoleExtensionOverrides(svc, &clusterRoleImpl{})
	schema.RegisterCoreV2ClusterRoleBinding(svc, &clusterRoleBindingImpl{})
	schema.RegisterCoreV2ClusterRoleBindingExtensionOverrides(svc, &clusterRoleBindingImpl{})
	schema.RegisterCoreV2Role(svc, &roleImpl{})
	schema.RegisterCoreV2RoleExtensionOverrides(svc, &roleImpl{})
	schema.RegisterCoreV2RoleBinding(svc, &roleBindingImpl{})
	schema.RegisterCoreV2RoleBindingExtensionOverrides(svc, &roleBindingImpl{})
	schema.RegisterCoreV2RoleRef(svc, &schema.CoreV2RoleRefAliases{})
	schema.RegisterCoreV2Rule(svc, &schema.CoreV2RuleAliases{})
	schema.RegisterCoreV2Subject(svc, &schema.CoreV2SubjectAliases{})

	// Register user types
	schema.RegisterUser(svc, &userImpl{})

	// Register version types
	schema.RegisterVersions(svc, &versionsImpl{})
	schema.RegisterEtcdVersions(svc, &schema.EtcdVersionsAliases{})
	schema.RegisterSensuBackendVersion(svc, &sensuBackendVersionImpl{})

	// Register mutations
	schema.RegisterMutation(svc, &mutationsImpl{svc: cfg})
	schema.RegisterCheckConfigInputs(svc)
	schema.RegisterCreateCheckInput(svc)
	schema.RegisterCreateCheckPayload(svc, &checkMutationPayload{})
	schema.RegisterCreateSilenceInput(svc)
	schema.RegisterCreateSilencePayload(svc, &schema.CreateSilencePayloadAliases{})
	schema.RegisterDeleteRecordInput(svc)
	schema.RegisterDeleteRecordPayload(svc, &deleteRecordPayload{})
	schema.RegisterExecuteCheckInput(svc)
	schema.RegisterExecuteCheckPayload(svc, &schema.ExecuteCheckPayloadAliases{})
	schema.RegisterResolveEventInput(svc)
	schema.RegisterSilenceInputs(svc)
	schema.RegisterUpdateCheckInput(svc)
	schema.RegisterUpdateCheckPayload(svc, &checkMutationPayload{})
	schema.RegisterPutWrappedPayload(svc, &schema.PutWrappedPayloadAliases{})

	// Errors
	schema.RegisterStandardError(svc, stdErrImpl{})
	schema.RegisterError(svc, &errImpl{})

	// Run init hooks allowing consumers to extend service
	for _, hookFn := range InitHooks {
		hookFn(svc, cfg)
	}

	// Configure tracing
	tracer := tracing.NewPrometheusTracer()
	tracer.AllowList = []string{
		tracing.KeyParse,
		tracing.KeyValidate,
		tracing.KeyExecuteQuery,
		tracing.KeyExecuteField,
	}
	svc.RegisterMiddleware(tracer)

	err := svc.Regenerate()
	return &wrapper, err
}

// Do executes given query string and variables
func (svc *Service) Do(ctx context.Context, p graphql.QueryParams) *graphql.Result {
	// Instantiate loaders and lift them into the context
	qryCtx := contextWithLoaders(ctx, *svc.Config)

	// Execute query inside context
	return svc.Target.Do(qryCtx, p)
}

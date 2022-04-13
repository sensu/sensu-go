package v2

// automatically generated file, do not edit!

import (
	"testing"
)

func TestResolveAPIKey(t *testing.T) {
	var value interface{} = new(APIKey)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("APIKey"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("APIKey")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"APIKey" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveAdhocRequest(t *testing.T) {
	var value interface{} = new(AdhocRequest)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("AdhocRequest"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("AdhocRequest")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"AdhocRequest" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveAny(t *testing.T) {
	var value interface{} = new(Any)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("Any"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("Any")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"Any" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveAsset(t *testing.T) {
	var value interface{} = new(Asset)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("Asset"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("Asset")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"Asset" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveAssetBuild(t *testing.T) {
	var value interface{} = new(AssetBuild)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("AssetBuild"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("AssetBuild")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"AssetBuild" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveAssetList(t *testing.T) {
	var value interface{} = new(AssetList)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("AssetList"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("AssetList")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"AssetList" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveAuthProviderClaims(t *testing.T) {
	var value interface{} = new(AuthProviderClaims)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("AuthProviderClaims"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("AuthProviderClaims")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"AuthProviderClaims" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveCheck(t *testing.T) {
	var value interface{} = new(Check)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("Check"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("Check")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"Check" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveCheckConfig(t *testing.T) {
	var value interface{} = new(CheckConfig)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("CheckConfig"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("CheckConfig")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"CheckConfig" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveCheckHistory(t *testing.T) {
	var value interface{} = new(CheckHistory)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("CheckHistory"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("CheckHistory")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"CheckHistory" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveCheckRequest(t *testing.T) {
	var value interface{} = new(CheckRequest)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("CheckRequest"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("CheckRequest")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"CheckRequest" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveClaims(t *testing.T) {
	var value interface{} = new(Claims)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("Claims"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("Claims")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"Claims" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveClusterHealth(t *testing.T) {
	var value interface{} = new(ClusterHealth)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("ClusterHealth"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("ClusterHealth")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"ClusterHealth" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveClusterRole(t *testing.T) {
	var value interface{} = new(ClusterRole)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("ClusterRole"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("ClusterRole")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"ClusterRole" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveClusterRoleBinding(t *testing.T) {
	var value interface{} = new(ClusterRoleBinding)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("ClusterRoleBinding"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("ClusterRoleBinding")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"ClusterRoleBinding" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveDeregistration(t *testing.T) {
	var value interface{} = new(Deregistration)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("Deregistration"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("Deregistration")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"Deregistration" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveEntity(t *testing.T) {
	var value interface{} = new(Entity)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("Entity"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("Entity")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"Entity" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveEvent(t *testing.T) {
	var value interface{} = new(Event)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("Event"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("Event")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"Event" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveEventFilter(t *testing.T) {
	var value interface{} = new(EventFilter)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("EventFilter"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("EventFilter")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"EventFilter" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveExtension(t *testing.T) {
	var value interface{} = new(Extension)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("Extension"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("Extension")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"Extension" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveHandler(t *testing.T) {
	var value interface{} = new(Handler)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("Handler"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("Handler")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"Handler" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveHandlerSocket(t *testing.T) {
	var value interface{} = new(HandlerSocket)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("HandlerSocket"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("HandlerSocket")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"HandlerSocket" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveHealthResponse(t *testing.T) {
	var value interface{} = new(HealthResponse)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("HealthResponse"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("HealthResponse")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"HealthResponse" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveHook(t *testing.T) {
	var value interface{} = new(Hook)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("Hook"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("Hook")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"Hook" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveHookConfig(t *testing.T) {
	var value interface{} = new(HookConfig)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("HookConfig"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("HookConfig")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"HookConfig" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveHookList(t *testing.T) {
	var value interface{} = new(HookList)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("HookList"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("HookList")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"HookList" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveKeepaliveRecord(t *testing.T) {
	var value interface{} = new(KeepaliveRecord)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("KeepaliveRecord"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("KeepaliveRecord")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"KeepaliveRecord" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveMetricPoint(t *testing.T) {
	var value interface{} = new(MetricPoint)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("MetricPoint"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("MetricPoint")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"MetricPoint" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveMetricTag(t *testing.T) {
	var value interface{} = new(MetricTag)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("MetricTag"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("MetricTag")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"MetricTag" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveMetricThreshold(t *testing.T) {
	var value interface{} = new(MetricThreshold)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("MetricThreshold"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("MetricThreshold")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"MetricThreshold" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveMetricThresholdRule(t *testing.T) {
	var value interface{} = new(MetricThresholdRule)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("MetricThresholdRule"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("MetricThresholdRule")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"MetricThresholdRule" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveMetricThresholdTag(t *testing.T) {
	var value interface{} = new(MetricThresholdTag)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("MetricThresholdTag"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("MetricThresholdTag")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"MetricThresholdTag" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveMetrics(t *testing.T) {
	var value interface{} = new(Metrics)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("Metrics"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("Metrics")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"Metrics" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveMutator(t *testing.T) {
	var value interface{} = new(Mutator)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("Mutator"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("Mutator")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"Mutator" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveNamespace(t *testing.T) {
	var value interface{} = new(Namespace)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("Namespace"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("Namespace")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"Namespace" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveNetwork(t *testing.T) {
	var value interface{} = new(Network)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("Network"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("Network")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"Network" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveNetworkInterface(t *testing.T) {
	var value interface{} = new(NetworkInterface)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("NetworkInterface"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("NetworkInterface")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"NetworkInterface" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveObjectMeta(t *testing.T) {
	var value interface{} = new(ObjectMeta)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("ObjectMeta"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("ObjectMeta")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"ObjectMeta" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolvePipeline(t *testing.T) {
	var value interface{} = new(Pipeline)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("Pipeline"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("Pipeline")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"Pipeline" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolvePipelineWorkflow(t *testing.T) {
	var value interface{} = new(PipelineWorkflow)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("PipelineWorkflow"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("PipelineWorkflow")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"PipelineWorkflow" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolvePostgresHealth(t *testing.T) {
	var value interface{} = new(PostgresHealth)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("PostgresHealth"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("PostgresHealth")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"PostgresHealth" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveProcess(t *testing.T) {
	var value interface{} = new(Process)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("Process"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("Process")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"Process" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveProxyRequests(t *testing.T) {
	var value interface{} = new(ProxyRequests)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("ProxyRequests"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("ProxyRequests")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"ProxyRequests" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveResourceReference(t *testing.T) {
	var value interface{} = new(ResourceReference)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("ResourceReference"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("ResourceReference")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"ResourceReference" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveRole(t *testing.T) {
	var value interface{} = new(Role)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("Role"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("Role")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"Role" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveRoleBinding(t *testing.T) {
	var value interface{} = new(RoleBinding)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("RoleBinding"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("RoleBinding")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"RoleBinding" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveRoleRef(t *testing.T) {
	var value interface{} = new(RoleRef)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("RoleRef"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("RoleRef")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"RoleRef" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveRule(t *testing.T) {
	var value interface{} = new(Rule)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("Rule"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("Rule")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"Rule" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveSecret(t *testing.T) {
	var value interface{} = new(Secret)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("Secret"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("Secret")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"Secret" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveSilenced(t *testing.T) {
	var value interface{} = new(Silenced)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("Silenced"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("Silenced")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"Silenced" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveSubject(t *testing.T) {
	var value interface{} = new(Subject)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("Subject"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("Subject")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"Subject" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveSystem(t *testing.T) {
	var value interface{} = new(System)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("System"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("System")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"System" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveTLSOptions(t *testing.T) {
	var value interface{} = new(TLSOptions)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("TLSOptions"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("TLSOptions")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"TLSOptions" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveTessenConfig(t *testing.T) {
	var value interface{} = new(TessenConfig)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("TessenConfig"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("TessenConfig")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"TessenConfig" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveTimeWindowDays(t *testing.T) {
	var value interface{} = new(TimeWindowDays)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("TimeWindowDays"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("TimeWindowDays")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"TimeWindowDays" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveTimeWindowRepeated(t *testing.T) {
	var value interface{} = new(TimeWindowRepeated)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("TimeWindowRepeated"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("TimeWindowRepeated")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"TimeWindowRepeated" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveTimeWindowTimeRange(t *testing.T) {
	var value interface{} = new(TimeWindowTimeRange)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("TimeWindowTimeRange"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("TimeWindowTimeRange")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"TimeWindowTimeRange" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveTimeWindowWhen(t *testing.T) {
	var value interface{} = new(TimeWindowWhen)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("TimeWindowWhen"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("TimeWindowWhen")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"TimeWindowWhen" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveTokens(t *testing.T) {
	var value interface{} = new(Tokens)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("Tokens"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("Tokens")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"Tokens" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveTypeMeta(t *testing.T) {
	var value interface{} = new(TypeMeta)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("TypeMeta"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("TypeMeta")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"TypeMeta" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveUser(t *testing.T) {
	var value interface{} = new(User)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("User"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("User")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"User" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveVersion(t *testing.T) {
	var value interface{} = new(Version)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("Version"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("Version")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"Version" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveNotExists(t *testing.T) {
	_, err := ResolveResource("!#$@$%@#$")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
}

package v2

// FixtureCheckRequest returns a fixture for a CheckRequest object.
func FixtureCheckRequest(id string) *CheckRequest {
	config := FixtureCheckConfig(id)

	return &CheckRequest{
		Config: config,
		Assets: []Asset{
			*FixtureAsset("ruby-2-4-2"),
		},
		Hooks: []HookConfig{
			*FixtureHookConfig("hook1"),
		},
	}
}

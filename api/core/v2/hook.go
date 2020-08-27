package v2

import (
	"errors"
	fmt "fmt"
	"net/url"
	"path"
	"regexp"
	"time"

	jsoniter "github.com/json-iterator/go"
	stringsutil "github.com/sensu/sensu-go/api/core/v2/internal/stringutil"
)

const (
	// HooksResource is the name of this resource type
	HooksResource = "hooks"

	// HookRequestType is the message type string for hook request.
	HookRequestType = "hook_request"
)

var (
	// CheckHookRegexStr used to validate type of check hook
	CheckHookRegexStr = `([0-9]|[1-9][0-9]|1[0-9][0-9]|2[0-4][0-9]|25[0-5])`

	// CheckHookRegex used to validate type of check hook
	CheckHookRegex = regexp.MustCompile("^" + CheckHookRegexStr + "$")

	// Severities used to validate type of check hook
	Severities = []string{"ok", "warning", "critical", "unknown", "non-zero"}
)

// Validate returns an error if the hook does not pass validation tests.
func (h *Hook) Validate() error {
	if err := h.HookConfig.Validate(); err != nil {
		return err
	}

	if h.Status < 0 {
		return errors.New("hook status must be greater than or equal to 0")
	}

	return nil
}

// StorePrefix returns the path prefix to this resource in the store
func (c *HookConfig) StorePrefix() string {
	return HooksResource
}

// URIPath returns the path component of a hook URI.
func (c *HookConfig) URIPath() string {
	if c.Namespace == "" {
		return path.Join(URLPrefix, HooksResource, url.PathEscape(c.Name))
	}
	return path.Join(URLPrefix, "namespaces", url.PathEscape(c.Namespace), HooksResource, url.PathEscape(c.Name))
}

// Validate returns an error if the hook does not pass validation tests.
func (c *HookConfig) Validate() error {
	if err := ValidateName(c.Name); err != nil {
		return errors.New("hook name " + err.Error())
	}

	if c.Command == "" {
		return errors.New("command cannot be empty")
	}

	if c.Timeout <= 0 {
		return errors.New("hook timeout must be greater than 0")
	}

	if c.Namespace == "" {
		return errors.New("namespace must be set")
	}

	return nil
}

// Validate returns an error if the check hook does not pass validation tests.
func (h *HookList) Validate() error {
	if h.Type == "" {
		return errors.New("type cannot be empty")
	}

	if h.Hooks == nil || len(h.Hooks) == 0 {
		return errors.New("hooks cannot be empty")
	}

	if !(CheckHookRegex.MatchString(h.Type) || isSeverity(h.Type)) {
		return errors.New(
			"valid check hook types are \"0\"-\"255\", \"ok\", \"warning\", \"critical\", \"unknown\", and \"non-zero\"",
		)
	}

	return nil
}

func isSeverity(name string) bool {
	for _, sev := range Severities {
		if sev == name {
			return true
		}
	}
	return false
}

// MarshalJSON implements the json.Marshaler interface.
func (h *HookList) MarshalJSON() ([]byte, error) {
	result := map[string][]string{h.Type: h.Hooks}
	return jsoniter.Marshal(result)
}

// UnmarshalJSON implements the json.Marshaler interface.
func (h *HookList) UnmarshalJSON(b []byte) error {
	result := map[string][]string{}
	if err := jsoniter.Unmarshal(b, &result); err != nil {
		return err
	}
	for k, v := range result {
		h.Type = k
		h.Hooks = v
	}

	return nil
}

// FixtureHookConfig returns a fixture for a HookConfig object.
func FixtureHookConfig(name string) *HookConfig {
	timeout := uint32(10)

	return &HookConfig{
		Command:    "true",
		Timeout:    timeout,
		Stdin:      false,
		ObjectMeta: NewObjectMeta(name, "default"),
	}
}

// FixtureHook returns a fixture for a Hook object.
func FixtureHook(id string) *Hook {
	t := time.Now().Unix()
	config := FixtureHookConfig(id)

	return &Hook{
		Status:     0,
		Output:     "",
		Issued:     t,
		Executed:   t + 1,
		Duration:   1.0,
		HookConfig: *config,
	}
}

// FixtureHookList returns a fixture for a HookList object.
func FixtureHookList(hookName string) *HookList {
	return &HookList{
		Hooks: []string{hookName},
		Type:  "non-zero",
	}
}

// URIPath returns the path component of a Hook URI.
func (h *Hook) URIPath() string {
	return fmt.Sprintf("/api/core/v2/namespaces/%s/hooks/%s", url.PathEscape(h.Namespace), url.PathEscape(h.Name))
}

// HookConfigFields returns a set of fields that represent that resource
func HookConfigFields(r Resource) map[string]string {
	resource := r.(*HookConfig)
	fields := map[string]string{
		"hook.name":      resource.ObjectMeta.Name,
		"hook.namespace": resource.ObjectMeta.Namespace,
	}
	stringsutil.MergeMapWithPrefix(fields, resource.ObjectMeta.Labels, "hook.labels.")
	return fields
}

// SetNamespace sets the namespace of the resource.
func (c *HookConfig) SetNamespace(namespace string) {
	c.Namespace = namespace
}

// SetObjectMeta sets the meta of the resource.
func (h *HookConfig) SetObjectMeta(meta ObjectMeta) {
	h.ObjectMeta = meta
}

// SetNamespace sets the namespace of the resource.
func (h *Hook) SetNamespace(namespace string) {
	h.Namespace = namespace
}

func (*Hook) RBACName() string {
	return "hooks"
}

func (*HookConfig) RBACName() string {
	return "hooks"
}

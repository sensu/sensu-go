package v2

// compatibility shims for core/v3.Resource support

func (a *APIKey) StoreName() string {
	return "api_keys"
}

func (a *APIKey) GetMetadata() *ObjectMeta {
	return &a.ObjectMeta
}

func (a *APIKey) SetMetadata(meta *ObjectMeta) {
	a.ObjectMeta = *meta
}

func (a *AdhocRequest) StoreName() string {
	return "adhoc_requests"
}

func (a *AdhocRequest) GetMetadata() *ObjectMeta {
	return &a.ObjectMeta
}

func (a *AdhocRequest) SetMetadata(meta *ObjectMeta) {
	a.ObjectMeta = *meta
}

func (a *Asset) StoreName() string {
	return "assets"
}

func (a *Asset) GetMetadata() *ObjectMeta {
	return &a.ObjectMeta
}

func (a *Asset) SetMetadata(meta *ObjectMeta) {
	a.ObjectMeta = *meta
}

func (c *Check) StoreName() string {
	return "checks"
}

func (c *Check) GetMetadata() *ObjectMeta {
	return &c.ObjectMeta
}

func (c *Check) SetMetadata(meta *ObjectMeta) {
	c.ObjectMeta = *meta
}

func (c *CheckConfig) StoreName() string {
	return "check_configs"
}

func (c *CheckConfig) GetMetadata() *ObjectMeta {
	return &c.ObjectMeta
}

func (c *CheckConfig) SetMetadata(meta *ObjectMeta) {
	c.ObjectMeta = *meta
}

func (c *ClusterRole) StoreName() string {
	return "cluster_roles"
}

func (c *ClusterRole) GetMetadata() *ObjectMeta {
	return &c.ObjectMeta
}

func (c *ClusterRole) SetMetadata(meta *ObjectMeta) {
	c.ObjectMeta = *meta
}

func (c *ClusterRoleBinding) StoreName() string {
	return "cluster_role_bindings"
}

func (c *ClusterRoleBinding) GetMetadata() *ObjectMeta {
	return &c.ObjectMeta
}

func (c *ClusterRoleBinding) SetMetadata(meta *ObjectMeta) {
	c.ObjectMeta = *meta
}

func (e *Entity) StoreName() string {
	return "entities"
}

func (e *Entity) GetMetadata() *ObjectMeta {
	return &e.ObjectMeta
}

func (e *Entity) SetMetadata(meta *ObjectMeta) {
	e.ObjectMeta = *meta
}

func (e *Event) StoreName() string {
	return "events"
}

func (e *Event) GetMetadata() *ObjectMeta {
	return &e.ObjectMeta
}

func (e *Event) SetMetadata(meta *ObjectMeta) {
	e.ObjectMeta = *meta
}

func (e *EventFilter) StoreName() string {
	return "event_filters"
}

func (e *EventFilter) GetMetadata() *ObjectMeta {
	return &e.ObjectMeta
}

func (e *EventFilter) SetMetadata(meta *ObjectMeta) {
	e.ObjectMeta = *meta
}

func (e *Extension) StoreName() string {
	return "extensions"
}

func (e *Extension) GetMetadata() *ObjectMeta {
	return &e.ObjectMeta
}

func (e *Extension) SetMetadata(meta *ObjectMeta) {
	e.ObjectMeta = *meta
}

func (h *Handler) StoreName() string {
	return "handlers"
}

func (h *Handler) GetMetadata() *ObjectMeta {
	return &h.ObjectMeta
}

func (h *Handler) SetMetadata(meta *ObjectMeta) {
	h.ObjectMeta = *meta
}

func (h *Hook) StoreName() string {
	return "hooks"
}

func (h *Hook) GetMetadata() *ObjectMeta {
	return &h.ObjectMeta
}

func (h *Hook) SetMetadata(meta *ObjectMeta) {
	h.ObjectMeta = *meta
}

func (h *HookConfig) StoreName() string {
	return "hook_configs"
}

func (h *HookConfig) GetMetadata() *ObjectMeta {
	return &h.ObjectMeta
}

func (h *HookConfig) SetMetadata(meta *ObjectMeta) {
	h.ObjectMeta = *meta
}

func (m *Mutator) StoreName() string {
	return "mutators"
}

func (m *Mutator) GetMetadata() *ObjectMeta {
	return &m.ObjectMeta
}

func (m *Mutator) SetMetadata(meta *ObjectMeta) {
	m.ObjectMeta = *meta
}

func (p *Pipeline) StoreName() string {
	return "pipelines"
}

func (p *Pipeline) GetMetadata() *ObjectMeta {
	return &p.ObjectMeta
}

func (p *Pipeline) SetMetadata(meta *ObjectMeta) {
	p.ObjectMeta = *meta
}

func (r *Role) StoreName() string {
	return "roles"
}

func (r *Role) GetMetadata() *ObjectMeta {
	return &r.ObjectMeta
}

func (r *Role) SetMetadata(meta *ObjectMeta) {
	r.ObjectMeta = *meta
}

func (r *RoleBinding) StoreName() string {
	return "role_bindings"
}

func (r *RoleBinding) GetMetadata() *ObjectMeta {
	return &r.ObjectMeta
}

func (r *RoleBinding) SetMetadata(meta *ObjectMeta) {
	r.ObjectMeta = *meta
}

func (s *Silenced) StoreName() string {
	return "silenceds"
}

func (s *Silenced) GetMetadata() *ObjectMeta {
	return &s.ObjectMeta
}

func (s *Silenced) SetMetadata(meta *ObjectMeta) {
	s.ObjectMeta = *meta
}

func (u *User) StoreName() string {
	return "users"
}

// TODO: We may want to add metadata field to the User type.
func (u *User) GetMetadata() *ObjectMeta {
	meta := u.GetObjectMeta()
	return &meta
}

func (u *User) SetMetadata(meta *ObjectMeta) {
	u.SetObjectMeta(*meta)
}

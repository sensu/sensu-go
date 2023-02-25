package postgres

import _ "embed"

//go:embed createOrUpdateEvent.sql
var createOrUpdateEvent string

//go:embed getEventByEntityCheck.sql
var getEventByEntityCheck string

//go:embed deleteEvent.sql
var deleteEvent string

//go:embed getKeepaliveCountsByNamespaceQuery.sql
var getKeepaliveCountsByNamespaceQuery string

//go:embed getEventCountsByNamespaceQuery.sql
var getEventCountsByNamespaceQuery string

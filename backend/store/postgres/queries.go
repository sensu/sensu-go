package postgres

import _ "embed"

//go:embed createOrUpdateEvent.sql
var createOrUpdateEvent string

//go:embed getEventByEntityCheck.sql
var getEventByEntityCheck string

//go:embed deleteEvent.sql
var deleteEvent string

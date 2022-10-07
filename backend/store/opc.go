package store

import (
	"context"
	"encoding/json"
	"time"
)

// OperatorType is used to distinguish operators of different types.
// For instance, it's allowed to have two operators with the same name within
// a namespace, provided they are of a different type. For example, a namespace
// could contain a keepalive operator called "ketchup" and also a check TTL
// operator called "ketchup".
type OperatorType int

// OperatorKey holds the key fields of an operator, for identifying a unique
// operator by namespace, type and name.
type OperatorKey struct {
	// Namespace is the namespace of the operator.
	Namespace string

	// Type is the type of the operator.
	Type OperatorType

	// Name is the name of the operator.
	Name string
}

// OperatorState holds state information about an operator, including a
// reference to its controller.
type OperatorState struct {
	// Namespace is the namespace of the operator.
	Namespace string

	// Type is the type of the operator.
	Type OperatorType

	// Name is the name of the operator.
	Name string

	// Controller is an operator that is responsible for this operator. It is
	// not guaranteed to be non-nil.
	Controller *OperatorKey

	// CheckInTimeout is the amount of time that can pass before the operator
	// would be considered absent.
	CheckInTimeout time.Duration

	// Present indicates whether or not the operator is currently present and
	// accounted for.
	Present bool

	// LastUpdate indicates when the operator last checked in, or when the OPC
	// system noted it was still absent.
	LastUpdate time.Time

	// Metadata is the operator's metadata. It can be any arbitrary JSON data
	// structure, so it is kept as an encoded json.RawMessage.
	Metadata *json.RawMessage
}

const (
	// NullOperator is a non-existent operator type.
	NullOperator OperatorType = 0

	// AgentOperator is the operator type for agent entities.
	AgentOperator OperatorType = 1

	// BackendOperator is the operator type for backend entities.
	BackendOperator OperatorType = 2

	// CheckOperator is the operator type for check TTLs functionality.
	CheckOperator OperatorType = 3
)

// MonitorOperatorsRequest is a request to watch a specific operator space
// for updates, whether those updates are generated as repeated absence alerts
// or check ins. The OPC will send all error encountered for each poll to the
// supplied error handler, if it is non-nil.
//
// In a clustered execution setting, it's often useful to use a distinct
// controller for each node so that nodes do not conflict in their monitoring,
// which involves resetting the LastUpdate field whenever the current time
// minus the time last seen is greater than the CheckInTimeout.
type MonitorOperatorsRequest struct {
	// Type is the OperatorType.
	Type OperatorType

	// Namespace is the namespace to monitor. It is optional.
	Namespace string

	// Name is the name to monitor. It is optional.
	Name string

	// ControllerType is the operator type of the controller.
	ControllerType OperatorType

	// ControllerNamespace is the namespace of the controller.
	ControllerNamespace string

	// ControllerName is the name of the controller.
	ControllerName string

	// Every is the operator database polling interval.
	Every time.Duration

	// ErrorHandler is a function tha handles any errors generated in the
	// monitoring process.
	ErrorHandler func(error)
}

// OperatorMonitor monitors operators for updates or missed check-ins.
type OperatorMonitor interface {
	// MonitorOperators continuously watches all operators in the operator
	// space for updates.
	MonitorOperators(context.Context, MonitorOperatorsRequest) <-chan []OperatorState
}

// OperatorQueryer lets users query for operators.
type OperatorQueryer interface {
	QueryOperator(context.Context, OperatorKey) (OperatorState, error)
}

// OperatorConcierge is responsible for checking operators in and out.
type OperatorConcierge interface {
	// CheckIn checks in an operator. It can be for the first time or a
	// scheduled check in.
	CheckIn(context.Context, OperatorState) error

	// CheckOut checks out an operator. The OperatorMonitor will no longer
	// notice absences for the operator, until it checks in again.
	CheckOut(context.Context, OperatorKey) error
}

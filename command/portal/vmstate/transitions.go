package vmstate

// Transition describes a single allowed operation from a given state.
type Transition struct {
	// Operation is the Nova API action name (e.g. "stop", "resize").
	Operation string
	// Target is the vm_state the instance will be in after the operation
	// succeeds. When the state does not change (e.g. reboot on ACTIVE,
	// snapshot, backup) Target equals the source state.
	Target VMStateName
}

// TransitionMap is keyed by vm_state; each value is the slice of
// transitions that are valid from that state.
// Delete and Fault are included for every non-terminal state because
// baseState handles them unconditionally.
type TransitionMap map[VMStateName][]Transition

// Transitions is the authoritative, statically-declared map of all
// valid state transitions in the OpenStack Nova state machine.
// It is derived directly from the same rules encoded in the concrete
// state types and from the official Nova documentation.
var Transitions TransitionMap = TransitionMap{
	StateBuilding: {
		// {Operation: "building_complete", Target: StateActive}, // internal Nova compute event
		// {Operation: "delete", Target: StateDeleted},
		// {Operation: "fault", Target: StateError},
	},

	StateActive: {
		{Operation: "pause", Target: StatePaused},
		// {Operation: "suspend", Target: StateSuspended},
		// {Operation: "rescue", Target: StateRescued},
		// {Operation: "resize", Target: StateResized},
		{Operation: "shelve", Target: StateShelved},
		// {Operation: "shelve_offload", Target: StateShelvedOffloaded},
		// {Operation: "soft_delete", Target: StateSoftDeleted},
		// {Operation: "delete", Target: StateDeleted},
		// {Operation: "fault", Target: StateError},
		// operations that leave the state unchanged
		{Operation: "reboot", Target: StateActive},
		// {Operation: "rebuild", Target: StateActive},
		// {Operation: "set_admin_password", Target: StateActive},
		// {Operation: "snapshot", Target: StateActive},
		// {Operation: "backup", Target: StateActive},
		{Operation: "stop", Target: StateStopped},
	},

	StatePaused: {
		{Operation: "unpause", Target: StateActive},
		{Operation: "shelve", Target: StateShelved},
		{Operation: "shelve_offload", Target: StateShelvedOffloaded},
		// {Operation: "delete", Target: StateDeleted},
		// {Operation: "fault", Target: StateError},
	},

	StateSuspended: {
		// {Operation: "resume", Target: StateActive},
		{Operation: "shelve", Target: StateShelved},
		{Operation: "shelve_offload", Target: StateShelvedOffloaded},
		// {Operation: "delete", Target: StateDeleted},
		// {Operation: "fault", Target: StateError},
	},

	StateStopped: {
		{Operation: "start", Target: StateActive},
		// {Operation: "rescue", Target: StateRescued},
		// {Operation: "resize", Target: StateResized},
		{Operation: "pause", Target: StatePaused},
		{Operation: "shelve", Target: StateShelved},
		// {Operation: "shelve_offload", Target: StateShelvedOffloaded},
		{Operation: "soft_delete", Target: StateSoftDeleted},
		// {Operation: "delete", Target: StateDeleted},
		// {Operation: "fault", Target: StateError},
		// stateless
		{Operation: "reboot", Target: StateActive},
		// {Operation: "rebuild", Target: StateStopped},
		// {Operation: "suspend", Target: StateSuspended},
		// {Operation: "snapshot", Target: StateStopped},
		// {Operation: "backup", Target: StateStopped},
	},

	StateRescued: {
		// {Operation: "unrescue", Target: StateActive},
		{Operation: "stop", Target: StateStopped},
		{Operation: "pause", Target: StatePaused},
		{Operation: "reboot", Target: StateRescued},
		// {Operation: "delete", Target: StateDeleted},
		// {Operation: "fault", Target: StateError},
	},

	StateResized: {
		// {Operation: "confirm_resize", Target: StateActive},
		// {Operation: "revert_resize", Target: StateActive},
		// {Operation: "delete", Target: StateDeleted},
		// {Operation: "fault", Target: StateError},
	},

	StateShelved: {
		{Operation: "unshelve", Target: StateActive},
		{Operation: "shelve_offload", Target: StateShelvedOffloaded},
		// {Operation: "delete", Target: StateDeleted},
		// {Operation: "fault", Target: StateError},
	},

	StateShelvedOffloaded: {
		{Operation: "unshelve", Target: StateActive},
		// {Operation: "delete", Target: StateDeleted},
		// {Operation: "fault", Target: StateError},
	},

	StateSoftDeleted: {
		// {Operation: "restore", Target: StateActive},
		// {Operation: "force_delete", Target: StateDeleted},
		// {Operation: "fault", Target: StateError},
	},

	StateError: {
		// {Operation: "soft_delete", Target: StateSoftDeleted},
		// {Operation: "rebuild", Target: StateActive},
		// {Operation: "delete", Target: StateDeleted},
	},

	// DELETED is terminal: no outgoing transitions.
	StateDeleted: {},
}

// AllowedOperations returns the slice of Transition values valid from
// the given state, or nil if the state is unknown.
func AllowedOperations(state string) []Transition {
	return Transitions[VMStateName(state)]
}

// CanTransition reports whether operation op is allowed from state from.
func CanTransition(from string, op string) bool {
	for _, t := range Transitions[VMStateName(from)] {
		if t.Operation == op {
			return true
		}
	}
	return false
}

// TargetState returns the target vm_state for the given operation
// executed from state from, and whether the operation is valid.
func TargetState(from string, op string) (VMStateName, bool) {
	for _, t := range Transitions[VMStateName(from)] {
		if t.Operation == op {
			return t.Target, true
		}
	}
	return "", false
}

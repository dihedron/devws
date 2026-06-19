package vmstate

// VMStateName is the canonical name of a stable vm_state.
type VMStateName string

const (
	StateBuilding         VMStateName = "BUILDING"
	StateActive           VMStateName = "ACTIVE"
	StatePaused           VMStateName = "PAUSED"
	StateSuspended        VMStateName = "SUSPENDED"
	StateStopped          VMStateName = "STOPPED"
	StateRescued          VMStateName = "RESCUED"
	StateResized          VMStateName = "RESIZED"
	StateShelved          VMStateName = "SHELVED"
	StateShelvedOffloaded VMStateName = "SHELVED_OFFLOADED"
	StateSoftDeleted      VMStateName = "SOFT_DELETED"
	StateDeleted          VMStateName = "DELETED"
	StateError            VMStateName = "ERROR"
)

var vmStatesMap = []VMStateName{
	StateBuilding,
	StateActive,
	StatePaused,
	StateSuspended,
	StateStopped,
	StateRescued,
	StateResized,
	StateShelved,
	StateShelvedOffloaded,
	StateSoftDeleted,
	StateDeleted,
	StateError,
}

func VMStates() []VMStateName {
	return vmStatesMap
}

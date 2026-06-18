package portal

import (
	"fmt"
	"log/slog"

	"github.com/dihedron/devws/openstack"
)

type Policy struct{}

// can view all vms
func (p Policy) CanViewVmsAsAdmin(user *User) bool {
	slog.Debug("CanViewVmsAsAdmin", "user", user, "PermVmsView", PermVmsView)
	return HasPermission(user, PermVmsView)
}

// can view vm
func (p Policy) CanViewVm(user *User) bool {
	slog.Debug("CanViewVm", "user", user, "PermVmsView", PermVmView)
	return HasPermission(user, PermVmView)
}

// can view specific vm detail
func (p Policy) CanViewVmDetail(user *User, vm openstack.Workstation) bool {
	slog.Debug("CanViewVmDetail", "user", user, "PermVmsView", PermVmViewDetail, "vm Id", vm.ID)
	if p.CanViewVmsAsAdmin(user) {
		return true
	}
	result := false
	if len(*vm.Tags) == 0 {
		return result
	}
	for _, tag := range *vm.Tags {
		if tag == fmt.Sprintf("devws.owner=%s", user.ID) && HasPermission(user, PermVmViewDetail) {
			result = true
			break
		}
	}
	return result
}

// can start a vm
func (p Policy) CanStartVm(user *User, vm openstack.Workstation) bool {
	if p.CanViewVmsAsAdmin(user) {
		return true
	}
	result := false
	if len(*vm.Tags) == 0 {
		return result
	}
	for _, tag := range *vm.Tags {
		if tag == fmt.Sprintf("devws.owner=%s", user.ID) && HasPermission(user, PermVmStart) {
			result = true
			break
		}
	}
	return result
}

// can stop a vm
func (p Policy) CanStopVm(user *User, vm openstack.Workstation) bool {
	if p.CanViewVmsAsAdmin(user) {
		return true
	}
	result := false
	if len(*vm.Tags) == 0 {
		return result
	}
	for _, tag := range *vm.Tags {
		if tag == fmt.Sprintf("devws.owner=%s", user.ID) && HasPermission(user, PermVmStop) {
			result = true
			break
		}
	}
	return result
}

// can shelve a vm
func (p Policy) CanShelveVm(user *User, vm openstack.Workstation) bool {
	if p.CanViewVmsAsAdmin(user) {
		return true
	}
	result := false
	if len(*vm.Tags) == 0 {
		return result
	}
	for _, tag := range *vm.Tags {
		if tag == fmt.Sprintf("devws.owner=%s", user.ID) && HasPermission(user, PermVmShelve) {
			result = true
			break
		}
	}
	return result
}

// can unshelve a vm
func (p Policy) CanUnShelveVm(user *User, vm openstack.Workstation) bool {
	if p.CanViewVmsAsAdmin(user) {
		return true
	}
	result := false
	if len(*vm.Tags) == 0 {
		return result
	}
	for _, tag := range *vm.Tags {
		if tag == fmt.Sprintf("devws.owner=%s", user.ID) && HasPermission(user, PermVmUnshelve) {
			result = true
			break
		}
	}
	return result
}

// can reboot a vm
func (p Policy) CanRebootVm(user *User, vm openstack.Workstation) bool {
	if p.CanViewVmsAsAdmin(user) {
		return true
	}
	result := false
	if len(*vm.Tags) == 0 {
		return result
	}
	for _, tag := range *vm.Tags {
		if tag == fmt.Sprintf("devws.owner=%s", user.ID) && HasPermission(user, PermVmReboot) {
			result = true
			break
		}
	}
	return result
}

// can tag a vm
func (p Policy) CanTagVm(user *User) bool {
	return p.CanViewVmsAsAdmin(user)
}

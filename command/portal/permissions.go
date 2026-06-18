package portal

type DomainRole string

const (
	DomainRoleAdmin     DomainRole = "DEVWS_ADMIN"
	DomainRoleDeveloper DomainRole = "DEVWS_DEVELOPER"
)

type Permission string

const (
	PermVmsView      Permission = "vms.view"
	PermVmViewDetail Permission = "vm.viewDetail"
	PermVmView       Permission = "vm.view"
	PermVmStart      Permission = "vm.start"
	PermVmStop       Permission = "vm.stop"
	PermVmShelve     Permission = "vm.shelve"
	PermVmUnshelve   Permission = "vm.unshelve"
	PermVmReboot     Permission = "vm.reboot"
	PermVmTag        Permission = "vm.tag"
)

type Role struct {
	Name        string
	Permissions []Permission
}

var roles = map[DomainRole]Role{
	DomainRoleAdmin: {
		Name: "ADMIN",
		Permissions: []Permission{
			PermVmsView,
			PermVmView,
			PermVmViewDetail,
			PermVmStart,
			PermVmStop,
			PermVmShelve,
			PermVmUnshelve,
			PermVmReboot,
			PermVmTag,
		},
	},
	DomainRoleDeveloper: {
		Name: "DEVELOPER",
		Permissions: []Permission{
			PermVmView,
			PermVmViewDetail,
			PermVmStart,
			PermVmStop,
			PermVmShelve,
			PermVmUnshelve,
			PermVmReboot,
		},
	},
}

type User struct {
	ID    string
	Email string
	Roles []Role
	Data  any
}

func HasPermission(user *User, perm Permission) bool {
	for _, role := range user.Roles {
		for _, p := range role.Permissions {
			if p == perm {
				return true
			}
		}
	}

	return false
}

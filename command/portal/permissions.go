package portal

type DomainRole string

const (
	DomainRoleAdmin     DomainRole = "DEVWS_ADMIN"
	DomainRoleDeveloper DomainRole = "DEVWS_DEVELOPER"
)

type Permission int16

const (
	PermNone    Permission = 0
	PermVmsView Permission = 1 << iota
	PermVmViewDetail
	PermVmView
	PermVmStart
	PermVmStop
	PermVmShelve
	PermVmUnshelve
	PermVmReboot
	PermVmPause
	PermAll = PermVmsView | PermVmViewDetail | PermVmView | PermVmStart | PermVmStop | PermVmShelve | PermVmUnshelve | PermVmReboot | PermVmPause
)

var permNames = map[Permission]string{
	PermVmsView:      "vms.view",
	PermVmViewDetail: "vm.viewDetail",
	PermVmView:       "vm.view",
	PermVmStart:      "vm.start",
	PermVmStop:       "vm.stop",
	PermVmShelve:     "vm.shelve",
	PermVmUnshelve:   "vm.unshelve",
	PermVmReboot:     "vm.reboot",
	PermVmPause:      "vm.pause",
}

var permValues = func() map[string]Permission {
	inv := make(map[string]Permission, len(permNames))
	for k, v := range permNames {
		inv[v] = k
	}
	return inv
}()

// Has controlla se il flag è presente (AND bit a bit)
func (p Permission) Has(flag Permission) bool {
	return p&flag == flag
}

// IsValid controlla che tutti i bit attivi siano flag noti
func (p Permission) IsValid() bool {
	return p&^PermAll == 0
}

func PermissionFromString(key string) (Permission, bool) {
	p, ok := permValues[key]
	return p, ok
}

func PermissionFromInt16(v int16) (Permission, bool) {
	p := Permission(v)
	return p, p.IsValid()
}

type Role struct {
	Name        string     `json:"name"`
	Permissions Permission `json:"permissions"`
}

var roles = map[DomainRole]Role{
	DomainRoleAdmin: {
		Name:        "ADMIN",
		Permissions: PermAll,
	},
	DomainRoleDeveloper: {
		Name:        "DEVELOPER",
		Permissions: PermVmView | PermVmViewDetail | PermVmStart | PermVmStop | PermVmShelve | PermVmUnshelve | PermVmReboot | PermVmPause,
	},
}

type UserSession struct {
	ID    string `json:"id"`
	Roles []Role `json:"roles"`
}

func HasPermission(user *UserSession, perm Permission) bool {
	for _, role := range user.Roles {
		if role.Permissions.Has(perm) {
			return true
		}
	}

	return false
}

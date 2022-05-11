package data

// RolesData is the structure that will keep information about ESDT roles
type RolesData map[string][]*RoleData

// RoleData is the structures that will keep information about a role
type RoleData struct {
	Token   string
	Address string
	Set     bool
}

// Add will add roles for the provided address
func (rd RolesData) Add(token string, address string, role string, set bool) {
	rData := &RoleData{
		Set:     set,
		Address: address,
		Token:   token,
	}

	_, found := rd[role]
	if found {
		rd[role] = append(rd[role], rData)
		return
	}

	rd[role] = []*RoleData{rData}
}

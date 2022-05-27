package data

// TokenRolesAndPropertiesHandler defines what a tokens and properties handler should do
type TokenRolesAndPropertiesHandler interface {
	AddRole(token string, address string, role string, set bool)
	AddProperties(token string, properties map[string]bool)
	GetRoles() map[string][]*RoleData
	GetAllTokensWithProperties() []*PropertiesData
}

type tokenRolesAndProperties struct {
	rolesData       map[string][]*RoleData
	tokenProperties []*PropertiesData
}

// RoleData is the structure that will keep information about a role
type RoleData struct {
	Token   string
	Address string
	Set     bool
}

// PropertiesData is the structure that will keep information about a token and properties
type PropertiesData struct {
	Token      string
	Properties map[string]bool
}

// NewTokenRolesAndProperties will create a new instance of tokenRolesAndProperties
// this is a NOT concurrent save structure
func NewTokenRolesAndProperties() *tokenRolesAndProperties {
	return &tokenRolesAndProperties{
		rolesData:       make(map[string][]*RoleData),
		tokenProperties: make([]*PropertiesData, 0),
	}
}

// AddRole will add role for the provided address
func (tap *tokenRolesAndProperties) AddRole(token string, address string, role string, set bool) {
	rData := &RoleData{
		Set:     set,
		Address: address,
		Token:   token,
	}

	_, found := tap.rolesData[role]
	if found {
		tap.rolesData[role] = append(tap.rolesData[role], rData)
		return
	}

	tap.rolesData[role] = []*RoleData{rData}
}

// GetRoles will return all the information about the roles
func (tap *tokenRolesAndProperties) GetRoles() map[string][]*RoleData {
	return tap.rolesData
}

// AddProperties will add token and the provided properties
func (tap *tokenRolesAndProperties) AddProperties(token string, properties map[string]bool) {
	tap.tokenProperties = append(tap.tokenProperties, &PropertiesData{
		Token:      token,
		Properties: properties,
	})
}

// GetAllTokensWithProperties will return all the tokens with properties
func (tap *tokenRolesAndProperties) GetAllTokensWithProperties() []*PropertiesData {
	return tap.tokenProperties
}

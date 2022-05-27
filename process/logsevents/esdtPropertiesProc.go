package logsevents

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
)

const (
	minTopicsPropertiesAndRoles = 4
	upgradePropertiesEvent      = "upgradeProperties"
)

type esdtPropertiesProc struct {
	pubKeyConverter            core.PubkeyConverter
	rolesOperationsIdentifiers map[string]struct{}
}

func newRolesProcessor(pubKeyConverter core.PubkeyConverter) *esdtPropertiesProc {
	return &esdtPropertiesProc{
		pubKeyConverter: pubKeyConverter,
		rolesOperationsIdentifiers: map[string]struct{}{
			core.BuiltInFunctionSetESDTRole:               {},
			core.BuiltInFunctionUnSetESDTRole:             {},
			core.BuiltInFunctionESDTNFTCreateRoleTransfer: {},
			upgradePropertiesEvent:                        {},
		},
	}
}

func (epp *esdtPropertiesProc) processEvent(args *argsProcessEvent) argOutputProcessEvent {
	identifier := string(args.event.GetIdentifier())
	_, ok := epp.rolesOperationsIdentifiers[identifier]
	if !ok {
		return argOutputProcessEvent{}
	}

	topics := args.event.GetTopics()
	if len(topics) < minTopicsPropertiesAndRoles {
		return argOutputProcessEvent{
			processed: true,
		}
	}

	if identifier == upgradePropertiesEvent {
		return epp.extractTokenProperties(args)
	}

	if identifier == core.BuiltInFunctionESDTNFTCreateRoleTransfer {
		return epp.extractDataNFTCreateRoleTransfer(args)
	}

	// topics contains:
	// [0] --> token identifier
	// [1] --> nonce of the NFT (bytes)
	// [2] --> value
	// [3:] --> roles to set or unset

	rolesBytes := topics[3:]
	shouldAddRole := identifier == core.BuiltInFunctionSetESDTRole
	addrBech := epp.pubKeyConverter.Encode(args.event.GetAddress())
	for _, roleBytes := range rolesBytes {
		args.tokenRolesAndProperties.AddRole(string(topics[0]), addrBech, string(roleBytes), shouldAddRole)
	}

	return argOutputProcessEvent{
		processed: true,
	}
}

func (epp *esdtPropertiesProc) extractDataNFTCreateRoleTransfer(args *argsProcessEvent) argOutputProcessEvent {
	topics := args.event.GetTopics()

	addrBech := epp.pubKeyConverter.Encode(args.event.GetAddress())
	shouldAddCreateRole := bytesToBool(topics[3])
	args.tokenRolesAndProperties.AddRole(string(topics[0]), addrBech, core.ESDTRoleNFTCreate, shouldAddCreateRole)

	return argOutputProcessEvent{
		processed: true,
	}
}

func (epp *esdtPropertiesProc) extractTokenProperties(args *argsProcessEvent) argOutputProcessEvent {
	topics := args.event.GetTopics()
	properties := topics[2:]
	propertiesMap := make(map[string]bool)
	for i := 0; i < len(properties); i += 2 {
		property := string(properties[i])
		val := bytesToBool(properties[i+1])
		propertiesMap[property] = val
	}

	args.tokenRolesAndProperties.AddProperties(string(topics[0]), propertiesMap)

	return argOutputProcessEvent{
		processed: true,
	}
}

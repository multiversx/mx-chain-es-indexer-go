package logsevents

import "github.com/ElrondNetwork/elrond-go-core/core"

type rolesProcessor struct {
	pubKeyConverter            core.PubkeyConverter
	rolesOperationsIdentifiers map[string]struct{}
}

func newRolesProcessor(pubKeyConverter core.PubkeyConverter) *rolesProcessor {
	return &rolesProcessor{
		pubKeyConverter: pubKeyConverter,
		rolesOperationsIdentifiers: map[string]struct{}{
			core.BuiltInFunctionSetESDTRole:               {},
			core.BuiltInFunctionUnSetESDTRole:             {},
			core.BuiltInFunctionESDTNFTCreateRoleTransfer: {},
		},
	}
}

func (rp *rolesProcessor) processEvent(args *argsProcessEvent) argOutputProcessEvent {
	identifier := string(args.event.GetIdentifier())
	_, ok := rp.rolesOperationsIdentifiers[identifier]
	if !ok {
		return argOutputProcessEvent{}
	}

	if identifier == core.BuiltInFunctionESDTNFTCreateRoleTransfer {
		return rp.extractDataNFTCreateRoleTransfer(args)
	}

	// topics contains:
	// [0] --> token identifier
	// [1] --> nonce of the NFT (bytes)
	// [2] --> value
	// [3:] --> roles to set or unset

	topics := args.event.GetTopics()
	if len(topics) < 4 {
		return argOutputProcessEvent{
			processed: true,
		}
	}

	rolesBytes := topics[3:]
	shouldAddRole := identifier == core.BuiltInFunctionSetESDTRole
	addrBech := rp.pubKeyConverter.Encode(args.event.GetAddress())
	for _, roleBytes := range rolesBytes {
		args.rolesData.Add(string(topics[0]), addrBech, string(roleBytes), shouldAddRole)
	}

	return argOutputProcessEvent{
		processed: true,
	}
}

func (rp *rolesProcessor) extractDataNFTCreateRoleTransfer(args *argsProcessEvent) argOutputProcessEvent {
	topics := args.event.GetTopics()
	if len(topics) < 4 {
		return argOutputProcessEvent{
			processed: true,
		}
	}

	addrBech := rp.pubKeyConverter.Encode(args.event.GetAddress())
	shouldAddCreateRole := bytesToBool(topics[3])
	args.rolesData.Add(string(topics[0]), addrBech, core.ESDTRoleNFTCreate, shouldAddCreateRole)

	return argOutputProcessEvent{
		processed: true,
	}
}

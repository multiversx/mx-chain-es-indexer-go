package wsindexer

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/stateChange"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/stretchr/testify/require"
)

// TestProcessPayload_MainBranchStateAccessesCausesWireTypeError reproduces the bug where
// mx-chain-go (main branch) sends OutportBlock with field 13 as:
//
//	map<string, StateAccesses> StateAccesses = 13
//
// But the indexer (feat/supernova-async-exec branch) expects field 13 as:
//
//	map<string, StateAccessesForBlock> StateAccessesForBlock = 13
//
// This mismatch causes "proto: illegal wireType 7" error during unmarshalling,
// specifically on start-of-epoch blocks where StateAccesses is populated.
func TestProcessPayload_MainBranchStateAccessesCausesWireTypeError(t *testing.T) {
	t.Parallel()

	marshaller := &marshal.GogoProtoMarshalizer{}

	// Simulate what mx-chain-go MAIN branch sends:
	// The main branch uses a DIFFERENT protobuf structure for field 13.
	// In main: map<string, StateAccesses> StateAccesses = 13
	// The StateAccesses message directly contains repeated StateAccess.
	//
	// We'll manually construct bytes that represent the MAIN branch format
	// to simulate receiving data from mx-chain-go main branch.

	// This is a minimal OutportBlock as sent by MAIN branch with populated StateAccesses
	// The key difference is field 13's value type.
	//
	// In MAIN branch, field 13 is: map<string, StateAccesses>
	// where StateAccesses = { repeated StateAccess StateAccess = 1 }
	//
	// In SUPERNOVA branch (indexer), field 13 is: map<string, StateAccessesForBlock>
	// where StateAccessesForBlock = { map<string, StateAccesses> StateAccesses = 1 }
	//
	// When main sends StateAccesses directly as the map value, and the indexer
	// tries to parse it as StateAccessesForBlock, the nested structure doesn't match.

	// Create an OutportBlock with populated StateAccessesForBlock (supernova format)
	// to first verify marshalling works correctly with matching types
	outportBlockSupernova := &outport.OutportBlock{
		ShardID:        0,
		NumberOfShards: 3,
		StateAccessesForBlock: map[string]*outport.StateAccessesForBlock{
			"txHash1": {
				StateAccesses: map[string]*stateChange.StateAccesses{
					"address1": {
						StateAccess: []*stateChange.StateAccess{
							{
								Type:        stateChange.Write,
								MainTrieKey: []byte("key1"),
								MainTrieVal: []byte("val1"),
							},
						},
					},
				},
			},
		},
	}

	// Marshal with supernova format - this should work fine
	supernovaBytes, err := marshaller.Marshal(outportBlockSupernova)
	require.NoError(t, err)
	require.NotEmpty(t, supernovaBytes)

	// Unmarshal back - this should work fine (same format)
	var unmarshalledSupernova outport.OutportBlock
	err = marshaller.Unmarshal(&unmarshalledSupernova, supernovaBytes)
	require.NoError(t, err)
	require.Equal(t, uint32(0), unmarshalledSupernova.ShardID)

	// Now simulate MAIN branch format by manually constructing incompatible bytes.
	// The MAIN branch sends field 13 with StateAccesses directly (not wrapped in StateAccessesForBlock).
	//
	// To simulate this, we construct bytes where field 13's map value is a StateAccesses
	// message (with field 1 = repeated StateAccess), but the indexer expects
	// StateAccessesForBlock (with field 1 = map<string, StateAccesses>).
	//
	// The structural difference causes wire type mismatch during parsing.

	// Construct a payload that mimics main branch encoding:
	// Field 13 in protobuf maps are encoded as: repeated MapEntry where MapEntry = {key, value}
	// The value in main is StateAccesses, but indexer expects StateAccessesForBlock.
	//
	// StateAccesses has: repeated StateAccess StateAccess = 1 (wire type 2, length-delimited)
	// StateAccessesForBlock has: map<string, StateAccesses> StateAccesses = 1 (wire type 2, but different internal structure)
	//
	// When the lengths and field numbers don't align, we get "illegal wireType" errors.

	// Create bytes that represent what MAIN branch would send
	// We'll create a minimal OutportBlock and then manipulate it to simulate the mismatch
	mainBranchBytes := createMainBranchOutportBlockBytes(t, marshaller)

	// Try to unmarshal main branch bytes with supernova's OutportBlock definition
	var unmarshalledFromMain outport.OutportBlock
	err = marshaller.Unmarshal(&unmarshalledFromMain, mainBranchBytes)

	// This SHOULD succeed without error if the protobuf structures are compatible.
	// The test FAILS (demonstrates the bug) if we get a protobuf parsing error,
	// which proves that main branch data cannot be parsed by supernova indexer.
	//
	// When this test fails, it means there's an incompatibility between:
	// - mx-chain-go (main branch) sending OutportBlock with StateAccesses
	// - mx-chain-es-indexer-go (supernova branch) expecting StateAccessesForBlock
	require.NoError(t, err, "BUG REPRODUCED: Main branch OutportBlock cannot be unmarshalled by supernova indexer. "+
		"Field 13 has incompatible types: main uses map<string, StateAccesses>, supernova expects map<string, StateAccessesForBlock>. "+
		"Error: %v", err)

	// If we get here without error, verify the data was correctly parsed
	require.NotNil(t, unmarshalledFromMain.StateAccessesForBlock, "StateAccessesForBlock should not be nil")
	require.Len(t, unmarshalledFromMain.StateAccessesForBlock, 1, "Should have 1 entry in StateAccessesForBlock")
}

// createMainBranchOutportBlockBytes creates bytes that simulate what mx-chain-go main branch
// would send. The main branch has a different protobuf structure for field 13.
//
// Main branch field 13: map<string, StateAccesses> StateAccesses = 13
// Supernova field 13:   map<string, StateAccessesForBlock> StateAccessesForBlock = 13
//
// The difference is that StateAccesses contains:
//
//	repeated StateAccess StateAccess = 1
//
// While StateAccessesForBlock contains:
//
//	map<string, StateAccesses> StateAccesses = 1
//
// This function manually constructs protobuf bytes that represent the main branch format.
func createMainBranchOutportBlockBytes(t *testing.T, marshaller *marshal.GogoProtoMarshalizer) []byte {
	// First, create the StateAccesses message as main branch would have it
	stateAccesses := &stateChange.StateAccesses{
		StateAccess: []*stateChange.StateAccess{
			{
				Type:        stateChange.Write,
				Index:       1,
				TxHash:      []byte("txhash123"),
				MainTrieKey: []byte("accountKey"),
				MainTrieVal: []byte("accountValue"),
				Operation:   1,
			},
		},
	}

	stateAccessesBytes, err := marshaller.Marshal(stateAccesses)
	require.NoError(t, err)

	// Now we need to construct an OutportBlock where field 13 contains
	// a map with StateAccesses as value (main branch format) instead of
	// StateAccessesForBlock (supernova format).
	//
	// In protobuf, a map<string, Message> is encoded as repeated entries where each entry
	// is a message with field 1 = key (string) and field 2 = value (message).
	//
	// For field 13 of OutportBlock:
	// - Field tag for map entry: (13 << 3) | 2 = 106 (wire type 2 = length-delimited)
	// - Each map entry contains: field 1 = key string, field 2 = value message
	//
	// Main branch: value is StateAccesses (has field 1 = repeated StateAccess)
	// Supernova: value is StateAccessesForBlock (has field 1 = map<string, StateAccesses>)

	// Build the map entry for main branch format
	// Map entry: { string key = 1; StateAccesses value = 2; }
	mapEntryKey := "testTxHash"
	mapEntryKeyBytes := encodeString(1, mapEntryKey)         // field 1, wire type 2
	mapEntryValueBytes := encodeBytes(2, stateAccessesBytes) // field 2, wire type 2
	mapEntryBytes := append(mapEntryKeyBytes, mapEntryValueBytes...)

	// Build OutportBlock with basic fields + field 13 (the problematic field)
	var outportBlockBytes []byte

	// Field 1: ShardID (uint32) = 0
	outportBlockBytes = append(outportBlockBytes, encodeVarint(1, 0)...)

	// Field 7: NumberOfShards (uint32) = 3
	outportBlockBytes = append(outportBlockBytes, encodeVarint(7, 3)...)

	// Field 13: StateAccesses map (main branch format)
	// This is where the incompatibility occurs
	outportBlockBytes = append(outportBlockBytes, encodeBytes(13, mapEntryBytes)...)

	return outportBlockBytes
}

// encodeVarint encodes a field with varint wire type (wire type 0)
func encodeVarint(fieldNumber int, value uint64) []byte {
	tag := uint64(fieldNumber<<3 | 0) // wire type 0 = varint
	var result []byte
	result = appendVarint(result, tag)
	result = appendVarint(result, value)
	return result
}

// encodeString encodes a string field (wire type 2)
func encodeString(fieldNumber int, value string) []byte {
	return encodeBytes(fieldNumber, []byte(value))
}

// encodeBytes encodes a bytes/message field (wire type 2)
func encodeBytes(fieldNumber int, value []byte) []byte {
	tag := uint64(fieldNumber<<3 | 2) // wire type 2 = length-delimited
	var result []byte
	result = appendVarint(result, tag)
	result = appendVarint(result, uint64(len(value)))
	result = append(result, value...)
	return result
}

// appendVarint appends a varint-encoded uint64 to the buffer
func appendVarint(buf []byte, v uint64) []byte {
	for v >= 0x80 {
		buf = append(buf, byte(v)|0x80)
		v >>= 7
	}
	buf = append(buf, byte(v))
	return buf
}

// TestProcessPayload_VerifyStateAccessesFieldDifference documents the structural difference
// between main and supernova branches for field 13 of OutportBlock.
func TestProcessPayload_VerifyStateAccessesFieldDifference(t *testing.T) {
	t.Parallel()

	// This test documents the protobuf structure difference:
	//
	// MAIN BRANCH (mx-chain-go master):
	// message OutportBlock {
	//   ...
	//   map<string, StateAccesses> StateAccesses = 13;
	// }
	// message StateAccesses {
	//   repeated StateAccess StateAccess = 1;
	// }
	//
	// SUPERNOVA BRANCH (indexer feat/supernova-async-exec):
	// message OutportBlock {
	//   ...
	//   map<string, StateAccessesForBlock> StateAccessesForBlock = 13;
	// }
	// message StateAccessesForBlock {
	//   map<string, StateAccesses> StateAccesses = 1;
	// }
	// message StateAccesses {
	//   repeated StateAccess StateAccess = 1;
	// }
	//
	// The difference is that supernova adds an extra wrapper (StateAccessesForBlock)
	// which changes the nested structure of the protobuf encoding.
	//
	// When mx-chain-go (main) sends data with the old structure, and the indexer
	// (supernova) tries to parse it with the new structure, the wire types don't match
	// because:
	// - Main's field 13 value has field 1 = repeated StateAccess (wire type 2, repeated message)
	// - Supernova's field 13 value expects field 1 = map<string, StateAccesses> (wire type 2, but map encoding)
	//
	// Maps in protobuf are encoded as repeated messages with {key, value} structure,
	// while repeated messages are just concatenated length-delimited messages.
	// This structural difference causes parsing errors.

	t.Log("Field 13 structure comparison:")
	t.Log("MAIN:     OutportBlock.StateAccesses[string] -> StateAccesses{repeated StateAccess}")
	t.Log("SUPERNOVA: OutportBlock.StateAccessesForBlock[string] -> StateAccessesForBlock{map[string]StateAccesses}")
	t.Log("")
	t.Log("The extra nesting level in SUPERNOVA causes wire type mismatch when parsing MAIN data")
}

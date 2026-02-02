package protocol

import (
	"encoding/json"
	"fmt"
)

// MustMarshal serializes a protocol message into JSON.
//
// This function panics on error because it is only intended to be used
// with server-owned, well-defined structs. A panic here indicates a
// programming error, not a runtime condition caused by client input.
func MustMarshal(message any) []byte {
	encoded, err := json.Marshal(message)
	if err != nil {
		panic(fmt.Sprintf("protocol marshal failed: %v", err))
	}
	return encoded
}

package protocol

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Protocol-level decode errors.
var (
	ErrInvalidJSON   = errors.New("invalid json")
	ErrMissingType   = errors.New(`missing "type" field`)
	ErrTypeNotString = errors.New(`"type" field is not a string`)
	ErrEmptyField    = errors.New("required field is empty")
)

// Envelope represents a minimally decoded message.
// It extracts the message type while preserving the raw JSON payload
// for strict, type-specific decoding.
type Envelope struct {
	Type MessageType
	Raw  json.RawMessage
}

// DecodeEnvelope parses a raw JSON frame and extracts the "type" field.
// The input must be a JSON object with a string-valued "type" field.
func DecodeEnvelope(frame []byte) (Envelope, error) {
	var decodedValue any
	if err := json.Unmarshal(frame, &decodedValue); err != nil {
		return Envelope{}, fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}

	objectMap, isObject := decodedValue.(map[string]any)
	if !isObject {
		return Envelope{}, fmt.Errorf("%w: expected json object", ErrInvalidJSON)
	}

	typeValue, exists := objectMap["type"]
	if !exists {
		return Envelope{}, ErrMissingType
	}

	typeString, isString := typeValue.(string)
	if !isString || typeString == "" {
		return Envelope{}, ErrTypeNotString
	}

	rawCopy := make([]byte, len(frame))
	copy(rawCopy, frame)

	return Envelope{
		Type: MessageType(typeString),
		Raw:  json.RawMessage(rawCopy),
	}, nil
}

// DecodeIdentify decodes and validates an IDENTIFY request.
func DecodeIdentify(envelope Envelope) (IdentifyRequest, error) {
	var request IdentifyRequest
	if err := json.Unmarshal(envelope.Raw, &request); err != nil {
		return IdentifyRequest{}, fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}

	if request.Type != TypeIdentify {
		return IdentifyRequest{}, fmt.Errorf(
			"expected message type %q, got %q",
			TypeIdentify,
			request.Type,
		)
	}

	if request.Username == "" {
		return IdentifyRequest{}, fmt.Errorf("%w: username", ErrEmptyField)
	}

	return request, nil
}

// DecodeStatus decodes and validates a STATUS request.
func DecodeStatus(envelope Envelope) (StatusRequest, error) {
	var request StatusRequest
	if err := json.Unmarshal(envelope.Raw, &request); err != nil {
		return StatusRequest{}, fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}

	if request.Type != TypeStatus {
		return StatusRequest{}, fmt.Errorf(
			"expected message type %q, got %q",
			TypeStatus,
			request.Type,
		)
	}

	switch request.Status {
	case StatusActive, StatusAway, StatusBusy:
		// valid
	default:
		return StatusRequest{}, fmt.Errorf("invalid status value: %q", request.Status)
	}

	return request, nil
}

// DecodeUsers decodes and validates a USERS request.
func DecodeUsers(envelope Envelope) (UsersRequest, error) {
	var request UsersRequest
	if err := json.Unmarshal(envelope.Raw, &request); err != nil {
		return UsersRequest{}, fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}

	if request.Type != TypeUsers {
		return UsersRequest{}, fmt.Errorf(
			"expected message type %q, got %q",
			TypeUsers,
			request.Type,
		)
	}

	return request, nil
}

// DecodeText decodes and validates a private TEXT request.
func DecodeText(envelope Envelope) (TextRequest, error) {
	var request TextRequest
	if err := json.Unmarshal(envelope.Raw, &request); err != nil {
		return TextRequest{}, fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}

	if request.Type != TypeText {
		return TextRequest{}, fmt.Errorf(
			"expected message type %q, got %q",
			TypeText,
			request.Type,
		)
	}

	if request.Username == "" {
		return TextRequest{}, fmt.Errorf("%w: username", ErrEmptyField)
	}
	if request.Text == "" {
		return TextRequest{}, fmt.Errorf("%w: text", ErrEmptyField)
	}

	return request, nil
}

// DecodePublicText decodes and validates a PUBLIC_TEXT request.
func DecodePublicText(envelope Envelope) (PublicTextRequest, error) {
	var request PublicTextRequest
	if err := json.Unmarshal(envelope.Raw, &request); err != nil {
		return PublicTextRequest{}, fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}

	if request.Type != TypePublicText {
		return PublicTextRequest{}, fmt.Errorf(
			"expected message type %q, got %q",
			TypePublicText,
			request.Type,
		)
	}

	if request.Text == "" {
		return PublicTextRequest{}, fmt.Errorf("%w: text", ErrEmptyField)
	}

	return request, nil
}

// DecodeNewRoom decodes and validates a NEW_ROOM request.
func DecodeNewRoom(envelope Envelope) (NewRoomRequest, error) {
	var request NewRoomRequest
	if err := json.Unmarshal(envelope.Raw, &request); err != nil {
		return NewRoomRequest{}, fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}

	if request.Type != TypeNewRoom {
		return NewRoomRequest{}, fmt.Errorf(
			"expected message type %q, got %q",
			TypeNewRoom,
			request.Type,
		)
	}

	if request.RoomName == "" {
		return NewRoomRequest{}, fmt.Errorf("%w: roomname", ErrEmptyField)
	}

	return request, nil
}

// DecodeInvite decodes and validates an INVITE request.
func DecodeInvite(envelope Envelope) (InviteRequest, error) {
	var request InviteRequest
	if err := json.Unmarshal(envelope.Raw, &request); err != nil {
		return InviteRequest{}, fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}

	if request.Type != TypeInvite {
		return InviteRequest{}, fmt.Errorf(
			"expected message type %q, got %q",
			TypeInvite,
			request.Type,
		)
	}

	if request.RoomName == "" {
		return InviteRequest{}, fmt.Errorf("%w: roomname", ErrEmptyField)
	}
	if len(request.Usernames) == 0 {
		return InviteRequest{}, fmt.Errorf("%w: usernames", ErrEmptyField)
	}

	for index, username := range request.Usernames {
		if username == "" {
			return InviteRequest{}, fmt.Errorf(
				"%w: usernames[%d]", ErrEmptyField, index,
			)
		}
	}

	return request, nil
}

// DecodeJoinRoom decodes and validates a JOIN_ROOM request.
func DecodeJoinRoom(envelope Envelope) (JoinRoomRequest, error) {
	var request JoinRoomRequest
	if err := json.Unmarshal(envelope.Raw, &request); err != nil {
		return JoinRoomRequest{}, fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}

	if request.Type != TypeJoinRoom {
		return JoinRoomRequest{}, fmt.Errorf(
			"expected message type %q, got %q",
			TypeJoinRoom,
			request.Type,
		)
	}

	if request.RoomName == "" {
		return JoinRoomRequest{}, fmt.Errorf("%w: roomname", ErrEmptyField)
	}

	return request, nil
}

// DecodeRoomUsers decodes and validates a ROOM_USERS request.
func DecodeRoomUsers(envelope Envelope) (RoomUsersRequest, error) {
	var request RoomUsersRequest
	if err := json.Unmarshal(envelope.Raw, &request); err != nil {
		return RoomUsersRequest{}, fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}

	if request.Type != TypeRoomUsers {
		return RoomUsersRequest{}, fmt.Errorf(
			"expected message type %q, got %q",
			TypeRoomUsers,
			request.Type,
		)
	}

	if request.RoomName == "" {
		return RoomUsersRequest{}, fmt.Errorf("%w: roomname", ErrEmptyField)
	}

	return request, nil
}

// DecodeRoomText decodes and validates a ROOM_TEXT request.
func DecodeRoomText(envelope Envelope) (RoomTextRequest, error) {
	var request RoomTextRequest
	if err := json.Unmarshal(envelope.Raw, &request); err != nil {
		return RoomTextRequest{}, fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}

	if request.Type != TypeRoomText {
		return RoomTextRequest{}, fmt.Errorf(
			"expected message type %q, got %q",
			TypeRoomText,
			request.Type,
		)
	}

	if request.RoomName == "" {
		return RoomTextRequest{}, fmt.Errorf("%w: roomname", ErrEmptyField)
	}
	if request.Text == "" {
		return RoomTextRequest{}, fmt.Errorf("%w: text", ErrEmptyField)
	}

	return request, nil
}

// DecodeLeaveRoom decodes and validates a LEAVE_ROOM request.
func DecodeLeaveRoom(envelope Envelope) (LeaveRoomRequest, error) {
	var request LeaveRoomRequest
	if err := json.Unmarshal(envelope.Raw, &request); err != nil {
		return LeaveRoomRequest{}, fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}

	if request.Type != TypeLeaveRoom {
		return LeaveRoomRequest{}, fmt.Errorf(
			"expected message type %q, got %q",
			TypeLeaveRoom,
			request.Type,
		)
	}

	if request.RoomName == "" {
		return LeaveRoomRequest{}, fmt.Errorf("%w: roomname", ErrEmptyField)
	}

	return request, nil
}

// DecodeDisconnect decodes and validates a DISCONNECT request.
func DecodeDisconnect(envelope Envelope) (DisconnectRequest, error) {
	var request DisconnectRequest
	if err := json.Unmarshal(envelope.Raw, &request); err != nil {
		return DisconnectRequest{}, fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}

	if request.Type != TypeDisconnect {
		return DisconnectRequest{}, fmt.Errorf(
			"expected message type %q, got %q",
			TypeDisconnect,
			request.Type,
		)
	}

	return request, nil
}

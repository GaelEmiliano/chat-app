package protocol

// Status represents a user's availability state
type Status string

const (
	StatusActive Status = "ACTIVE"
	StatusAway   Status = "AWAY"
	StatusBusy   Status = "BUSY"
)

// MessageType represents the value of the "type" field in all protocol messages
type MessageType string

const (
	// Client to Server
	TypeIdentify   MessageType = "IDENTIFY"
	TypeStatus     MessageType = "STATUS"
	TypeUsers      MessageType = "USERS"
	TypeText       MessageType = "TEXT"
	TypePublicText MessageType = "PUBLIC_TEXT"
	TypeNewRoom    MessageType = "NEW_ROOM"
	TypeInvite     MessageType = "INVITE"
	TypeJoinRoom   MessageType = "JOIN_ROOM"
	TypeRoomUsers  MessageType = "ROOM_USERS"
	TypeRoomText   MessageType = "ROOM_TEXT"
	TypeLeaveRoom  MessageType = "LEAVE_ROOM"
	TypeDisconnect MessageType = "DISCONNECT"

	// Server to Client
	TypeResponse       MessageType = "RESPONSE"
	TypeNewUser        MessageType = "NEW_USER"
	TypeNewStatus      MessageType = "NEW_STATUS"
	TypeUserList       MessageType = "USER_LIST"
	TypeTextFrom       MessageType = "TEXT_FROM"
	TypePublicTextFrom MessageType = "PUBLIC_TEXT_FROM"
	TypeInvitation     MessageType = "INVITATION"
	TypeJoinedRoom     MessageType = "JOINED_ROOM"
	TypeRoomUserList   MessageType = "ROOM_USER_LIST"
	TypeRoomTextFrom   MessageType = "ROOM_TEXT_FROM"
	TypeLeftRoom       MessageType = "LEFT_ROOM"
	TypeDisconnected   MessageType = "DISCONNECTED"
)

// Client to Server messages

// IdentifyRequest is sent by a client to identify itself when connecting.
type IdentifyRequest struct {
	Type     MessageType `json:"type"`
	Username string      `json:"username"`
}

// StatusRequest updates the user's status.
type StatusRequest struct {
	Type   MessageType `json:"type"`
	Status Status      `json:"status"`
}

// UsersRequest asks the server for the full user list and statuses.
type UsersRequest struct {
	Type MessageType `json:"type"`
}

// TextRequest sends a private message to a user.
type TextRequest struct {
	Type     MessageType `json:"type"`
	Username string      `json:"username"`
	Text     string      `json:"text"`
}

// PublicTextRequest sends a public message to all users except the sender.
type PublicTextRequest struct {
	Type MessageType `json:"type"`
	Text string      `json:"text"`
}

// NewRoomRequest creates a new room. The creator becomes the first member.
type NewRoomRequest struct {
	Type     MessageType `json:"type"`
	RoomName string      `json:"roomname"`
}

// InviteRequest invites users to a room.
// The server ignores users already invited or already joined.
type InviteRequest struct {
	Type      MessageType `json:"type"`
	RoomName  string      `json:"roomname"`
	Usernames []string    `json:"usernames"`
}

// JoinRoomRequest joins a room the user was invited to.
type JoinRoomRequest struct {
	Type     MessageType `json:"type"`
	RoomName string      `json:"roomname"`
}

// RoomUsersRequest asks for the list of users in a room.
type RoomUsersRequest struct {
	Type     MessageType `json:"type"`
	RoomName string      `json:"roomname"`
}

// RoomTextRequest sends a message to all users in a room except the sender.
type RoomTextRequest struct {
	Type     MessageType `json:"type"`
	RoomName string      `json:"roomname"`
	Text     string      `json:"text"`
}

// LeaveRoomRequest leaves a room the user previously joined.
type LeaveRoomRequest struct {
	Type     MessageType `json:"type"`
	RoomName string      `json:"roomname"`
}

// DisconnectRequest explicitly disconnects the client.
type DisconnectRequest struct {
	Type MessageType `json:"type"`
}

// Server to Client messages

// ResponseMessage is a generic server response for operations that require
// explicit acknowledgment or error reporting.
type ResponseMessage struct {
	Type      MessageType `json:"type"`
	Operation string      `json:"operation"`
	Result    string      `json:"result"`
	Extra     string      `json:"extra,omitempty"`
}

// NewUserMessage is broadcast when a new user successfully identifies.
type NewUserMessage struct {
	Type     MessageType `json:"type"`
	Username string      `json:"username"`
}

// NewStatusMessage is broadcast when a user changes status.
type NewStatusMessage struct {
	Type     MessageType `json:"type"`
	Username string      `json:"username"`
	Status   Status      `json:"status"`
}

// UserListMessage is sent in response to USERS.
type UserListMessage struct {
	Type  MessageType       `json:"type"`
	Users map[string]Status `json:"users"`
}

// TextFromMessage is delivered to a recipient for private messages.
type TextFromMessage struct {
	Type     MessageType `json:"type"`
	Username string      `json:"username"`
	Text     string      `json:"text"`
}

// PublicTextFromMessage is broadcast for public messages.
type PublicTextFromMessage struct {
	Type     MessageType `json:"type"`
	Username string      `json:"username"`
	Text     string      `json:"text"`
}

// InvitationMessage is sent to invited users.
type InvitationMessage struct {
	Type     MessageType `json:"type"`
	RoomName string      `json:"roomname"`
	Username string      `json:"username"`
}

// JoinedRoomMessage is broadcast to users in a room when someone joins.
type JoinedRoomMessage struct {
	Type     MessageType `json:"type"`
	RoomName string      `json:"roomname"`
	Username string      `json:"username"`
}

// RoomUserListMessage is sent in response to ROOM_USERS.
type RoomUserListMessage struct {
	Type     MessageType       `json:"type"`
	RoomName string            `json:"roomname"`
	Users    map[string]Status `json:"users"`
}

// RoomTextFromMessage is broadcast to room members for room messages.
type RoomTextFromMessage struct {
	Type     MessageType `json:"type"`
	RoomName string      `json:"roomname"`
	Username string      `json:"username"`
	Text     string      `json:"text"`
}

// LeftRoomMessage is broadcast to users in a room when someone leaves.
type LeftRoomMessage struct {
	Type     MessageType `json:"type"`
	RoomName string      `json:"roomname"`
	Username string      `json:"username"`
}

// DisconnectedMessage is broadcast when a user disconnects.
type DisconnectedMessage struct {
	Type     MessageType `json:"type"`
	Username string      `json:"username"`
}

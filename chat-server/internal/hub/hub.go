package hub

import (
	"context"
	"fmt"
	"log"
	"time"

	"chat-server/internal/config"
	"chat-server/internal/protocol"
)

// ClientID uniquely identifies a connected client within the server.
type ClientID string

// ClientWriter abstracts the outbound side of a client connection.
// The hub owns protocol decisions; the concrete client owns I/O.
type ClientWriter interface {
	Send(ctx context.Context, frame []byte) error
	Close() error
}

// InboundEvent represents a raw protocol frame received from a client.
type InboundEvent struct {
	ClientID ClientID
	Frame    []byte
	At       time.Time
}

// RegisterEvent registers a newly connected client with the hub.
type RegisterEvent struct {
	ClientID ClientID
	Writer   ClientWriter
}

// UnregisterEvent removes a client from the hub and triggers cleanup.
type UnregisterEvent struct {
	ClientID ClientID
	Reason   string
}

// ...
type RoomState struct {
	name    string
	members map[ClientID]struct{}
	invited map[ClientID]struct{}
}

// Hub is the single owner of all shared server state.
//
// Concurrency model:
//   - The hub runs in exactly one goroutine.
//   - All mutable state is accessed only inside Run().
//   - External goroutines communicate exclusively via channels.
//
// This design avoids locks and data races by construction.
type Hub struct {
	logger *log.Logger
	cfg    config.Config

	inbound    chan InboundEvent
	register   chan RegisterEvent
	unregister chan UnregisterEvent

	// State owned by the hub goroutine only.
	clients       map[ClientID]ClientWriter
	clientUser    map[ClientID]string
	clientStatus  map[ClientID]protocol.Status
	usernameOwner map[string]ClientID

	rooms       map[string]*RoomState
	clientRooms map[ClientID]map[string]struct{}
}

// New creates a new Hub instance.
// The caller must invoke Run() in its own goroutine.
func New(logger *log.Logger, cfg config.Config) *Hub {
	return &Hub{
		logger:        logger,
		cfg:           cfg,
		inbound:       make(chan InboundEvent, 256),
		register:      make(chan RegisterEvent, 256),
		unregister:    make(chan UnregisterEvent, 256),
		clients:       make(map[ClientID]ClientWriter),
		clientUser:    make(map[ClientID]string),
		clientStatus:  make(map[ClientID]protocol.Status),
		usernameOwner: make(map[string]ClientID),
		rooms:         make(map[string]*RoomState),
		clientRooms:   make(map[ClientID]map[string]struct{}),
	}
}

// Run processes all hub events until the context is canceled.
func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			h.closeAll("server shutting down")
			return

		case event := <-h.register:
			h.clients[event.ClientID] = event.Writer

		case event := <-h.unregister:
			h.forceDisconnect(ctx, event.ClientID, event.Reason)

		case event := <-h.inbound:
			h.handleInbound(ctx, event)
		}
	}
}

// Register registers a client connection with the hub.
func (h *Hub) Register(clientID ClientID, writer ClientWriter) {
	h.register <- RegisterEvent{
		ClientID: clientID,
		Writer:   writer,
	}
}

// Unregister requests removal of a client from the hub.
func (h *Hub) Unregister(clientID ClientID, reason string) {
	h.unregister <- UnregisterEvent{
		ClientID: clientID,
		Reason:   reason,
	}
}

// Deliver delivers a raw protocol frame from a client to the hub.
func (h *Hub) Deliver(clientID ClientID, frame []byte) {
	h.inbound <- InboundEvent{
		ClientID: clientID,
		Frame:    frame,
		At:       time.Now().UTC(),
	}
}

func (h *Hub) handleInbound(ctx context.Context, event InboundEvent) {
	envelope, err := protocol.DecodeEnvelope(event.Frame)
	if err != nil {
		h.sendInvalidAndDisconnect(ctx, event.ClientID, "INVALID", "INVALID")
		return
	}

	username, isIdentified := h.clientUser[event.ClientID]

	if !isIdentified {
		if envelope.Type != protocol.TypeIdentify {
			h.sendInvalidAndDisconnect(ctx, event.ClientID, "INVALID", "NOT_IDENTIFIED")
			return
		}
		h.handleIdentify(ctx, event.ClientID, envelope)
		return
	}

	switch envelope.Type {
	case protocol.TypeStatus:
		h.handleStatus(ctx, event.ClientID, username, envelope)

	case protocol.TypeUsers:
		h.handleUsers(ctx, event.ClientID, envelope)

	case protocol.TypeText:
		h.handleText(ctx, event.ClientID, username, envelope)

	case protocol.TypePublicText:
		h.handlePublicText(ctx, event.ClientID, username, envelope)

	case protocol.TypeNewRoom:
		h.handleNewRoom(ctx, event.ClientID, envelope)

	case protocol.TypeInvite:
		h.handleInvite(ctx, event.ClientID, username, envelope)

	case protocol.TypeJoinRoom:
		h.handleJoinRoom(ctx, event.ClientID, username, envelope)

	case protocol.TypeDisconnect:
		h.handleDisconnect(ctx, event.ClientID, username, envelope)

	case protocol.TypeRoomUsers:
		h.handleRoomUsers(ctx, event.ClientID, envelope)

	case protocol.TypeRoomText:
		h.handleRoomText(ctx, event.ClientID, username, envelope)

	case protocol.TypeLeaveRoom:
		h.handleLeaveRoom(ctx, event.ClientID, username, envelope)

	default:
		h.sendInvalidAndDisconnect(ctx, event.ClientID, "INVALID", "INVALID")
	}

}

func (h *Hub) handleIdentify(ctx context.Context, clientID ClientID, envelope protocol.Envelope) {
	request, err := protocol.DecodeIdentify(envelope)
	if err != nil {
		h.sendInvalidAndDisconnect(ctx, clientID, "INVALID", "INVALID")
		return
	}

	if len(request.Username) == 0 || len(request.Username) > h.cfg.MaxUsernameLength {
		h.sendInvalidAndDisconnect(ctx, clientID, "INVALID", "INVALID")
		return
	}

	if _, exists := h.usernameOwner[request.Username]; exists {
		h.sendResponse(ctx, clientID, protocol.ResponseMessage{
			Type:      protocol.TypeResponse,
			Operation: "IDENTIFY",
			Result:    "USER_ALREADY_EXISTS",
			Extra:     request.Username,
		})
		return
	}

	h.clientUser[clientID] = request.Username
	h.clientStatus[clientID] = protocol.StatusActive
	h.usernameOwner[request.Username] = clientID

	h.sendResponse(ctx, clientID, protocol.ResponseMessage{
		Type:      protocol.TypeResponse,
		Operation: "IDENTIFY",
		Result:    "SUCCESS",
		Extra:     request.Username,
	})

	h.broadcastExcept(ctx, clientID, protocol.MustMarshal(protocol.NewUserMessage{
		Type:     protocol.TypeNewUser,
		Username: request.Username,
	}))
}

func (h *Hub) handleStatus(
	ctx context.Context,
	clientID ClientID,
	username string,
	envelope protocol.Envelope,
) {
	request, err := protocol.DecodeStatus(envelope)
	if err != nil {
		h.sendInvalidAndDisconnect(ctx, clientID, "INVALID", "INVALID")
		return
	}

	h.clientStatus[clientID] = request.Status

	h.broadcastExcept(ctx, clientID, protocol.MustMarshal(protocol.NewStatusMessage{
		Type:     protocol.TypeNewStatus,
		Username: username,
		Status:   request.Status,
	}))
}

func (h *Hub) handleUsers(
	ctx context.Context,
	clientID ClientID,
	envelope protocol.Envelope,
) {
	_, err := protocol.DecodeUsers(envelope)
	if err != nil {
		h.sendInvalidAndDisconnect(ctx, clientID, "INVALID", "INVALID")
		return
	}

	usersSnapshot := make(map[string]protocol.Status, len(h.clientUser))
	for knownClientID, knownUsername := range h.clientUser {
		status, hasStatus := h.clientStatus[knownClientID]
		if !hasStatus {
			// Identified users should always have a status; default to ACTIVE defensively.
			status = protocol.StatusActive
		}
		usersSnapshot[knownUsername] = status
	}

	userListFrame := protocol.MustMarshal(protocol.UserListMessage{
		Type:  protocol.TypeUserList,
		Users: usersSnapshot,
	})

	h.sendFrame(ctx, clientID, userListFrame)
}

func (h *Hub) handleText(
	ctx context.Context,
	senderClientID ClientID,
	senderUsername string,
	envelope protocol.Envelope,
) {
	request, err := protocol.DecodeText(envelope)
	if err != nil {
		h.sendInvalidAndDisconnect(ctx, senderClientID, "INVALID", "INVALID")
		return
	}

	recipientClientID, exists := h.usernameOwner[request.Username]
	if !exists {
		h.sendResponse(ctx, senderClientID, protocol.ResponseMessage{
			Type:      protocol.TypeResponse,
			Operation: "TEXT",
			Result:    "NO_SUCH_USER",
			Extra:     request.Username,
		})
		return
	}

	textFrame := protocol.MustMarshal(protocol.TextFromMessage{
		Type:     protocol.TypeTextFrom,
		Username: senderUsername,
		Text:     request.Text,
	})

	h.sendFrame(ctx, recipientClientID, textFrame)
}

func (h *Hub) handlePublicText(
	ctx context.Context,
	senderClientID ClientID,
	senderUsername string,
	envelope protocol.Envelope,
) {
	request, err := protocol.DecodePublicText(envelope)
	if err != nil {
		h.sendInvalidAndDisconnect(ctx, senderClientID, "INVALID", "INVALID")
		return
	}

	publicTextFrame := protocol.MustMarshal(protocol.PublicTextFromMessage{
		Type:     protocol.TypePublicTextFrom,
		Username: senderUsername,
		Text:     request.Text,
	})

	h.broadcastExcept(ctx, senderClientID, publicTextFrame)
}

func (h *Hub) handleNewRoom(
	ctx context.Context,
	creatorClientID ClientID,
	envelope protocol.Envelope,
) {
	request, err := protocol.DecodeNewRoom(envelope)
	if err != nil {
		h.sendInvalidAndDisconnect(ctx, creatorClientID, "INVALID", "INVALID")
		return
	}

	if len(request.RoomName) == 0 || len(request.RoomName) > h.cfg.MaxRoomNameLength {
		h.sendInvalidAndDisconnect(ctx, creatorClientID, "INVALID", "INVALID")
		return
	}

	if _, exists := h.rooms[request.RoomName]; exists {
		h.sendResponse(ctx, creatorClientID, protocol.ResponseMessage{
			Type:      protocol.TypeResponse,
			Operation: "NEW_ROOM",
			Result:    "ROOM_ALREADY_EXISTS",
			Extra:     request.RoomName,
		})
		return
	}

	newRoom := &RoomState{
		name:    request.RoomName,
		members: make(map[ClientID]struct{}),
		invited: make(map[ClientID]struct{}),
	}
	newRoom.members[creatorClientID] = struct{}{}

	h.rooms[request.RoomName] = newRoom
	h.ensureClientRoomSet(creatorClientID)[request.RoomName] = struct{}{}

	h.sendResponse(ctx, creatorClientID, protocol.ResponseMessage{
		Type:      protocol.TypeResponse,
		Operation: "NEW_ROOM",
		Result:    "SUCCESS",
		Extra:     request.RoomName,
	})
}

func (h *Hub) handleInvite(
	ctx context.Context,
	inviterClientID ClientID,
	inviterUsername string,
	envelope protocol.Envelope,
) {
	request, err := protocol.DecodeInvite(envelope)
	if err != nil {
		h.sendInvalidAndDisconnect(ctx, inviterClientID, "INVALID", "INVALID")
		return
	}

	if len(request.RoomName) == 0 || len(request.RoomName) > h.cfg.MaxRoomNameLength {
		h.sendInvalidAndDisconnect(ctx, inviterClientID, "INVALID", "INVALID")
		return
	}

	room, exists := h.rooms[request.RoomName]
	if !exists {
		h.sendResponse(ctx, inviterClientID, protocol.ResponseMessage{
			Type:      protocol.TypeResponse,
			Operation: "INVITE",
			Result:    "NO_SUCH_ROOM",
			Extra:     request.RoomName,
		})
		return
	}

	// The spec states only users who are inside a room can invite others to that room.
	// This is treated as a protocol violation if the inviter is not a room member.
	if !h.isRoomMember(room, inviterClientID) {
		h.sendInvalidAndDisconnect(ctx, inviterClientID, "INVALID", "INVALID")
		return
	}

	recipientClientIDs := make([]ClientID, 0, len(request.Usernames))
	for _, targetUsername := range request.Usernames {
		targetClientID, userExists := h.usernameOwner[targetUsername]
		if !userExists {
			h.sendResponse(ctx, inviterClientID, protocol.ResponseMessage{
				Type:      protocol.TypeResponse,
				Operation: "INVITE",
				Result:    "NO_SUCH_USER",
				Extra:     targetUsername,
			})
			return
		}
		recipientClientIDs = append(recipientClientIDs, targetClientID)
	}

	invitationFrame := protocol.MustMarshal(protocol.InvitationMessage{
		Type:     protocol.TypeInvitation,
		RoomName: request.RoomName,
		Username: inviterUsername,
	})

	for _, recipientClientID := range recipientClientIDs {
		// Ignore already joined users.
		if _, isMember := room.members[recipientClientID]; isMember {
			continue
		}
		// Ignore already invited users.
		if _, alreadyInvited := room.invited[recipientClientID]; alreadyInvited {
			continue
		}

		room.invited[recipientClientID] = struct{}{}
		h.sendFrame(ctx, recipientClientID, invitationFrame)
	}
}

func (h *Hub) handleJoinRoom(
	ctx context.Context,
	clientID ClientID,
	username string,
	envelope protocol.Envelope,
) {
	request, err := protocol.DecodeJoinRoom(envelope)
	if err != nil {
		h.sendInvalidAndDisconnect(ctx, clientID, "INVALID", "INVALID")
		return
	}

	if len(request.RoomName) == 0 || len(request.RoomName) > h.cfg.MaxRoomNameLength {
		h.sendInvalidAndDisconnect(ctx, clientID, "INVALID", "INVALID")
		return
	}

	room, exists := h.rooms[request.RoomName]
	if !exists {
		h.sendResponse(ctx, clientID, protocol.ResponseMessage{
			Type:      protocol.TypeResponse,
			Operation: "JOIN_ROOM",
			Result:    "NO_SUCH_ROOM",
			Extra:     request.RoomName,
		})
		return
	}

	// Idempotency: if already a member, return SUCCESS without broadcasting again.
	if _, alreadyMember := room.members[clientID]; alreadyMember {
		h.sendResponse(ctx, clientID, protocol.ResponseMessage{
			Type:      protocol.TypeResponse,
			Operation: "JOIN_ROOM",
			Result:    "SUCCESS",
			Extra:     request.RoomName,
		})
		return
	}

	if _, wasInvited := room.invited[clientID]; !wasInvited {
		h.sendResponse(ctx, clientID, protocol.ResponseMessage{
			Type:      protocol.TypeResponse,
			Operation: "JOIN_ROOM",
			Result:    "NOT_INVITED",
			Extra:     request.RoomName,
		})
		return
	}

	// Transition: invited -> member
	delete(room.invited, clientID)
	room.members[clientID] = struct{}{}

	h.ensureClientRoomSet(clientID)[request.RoomName] = struct{}{}

	h.sendResponse(ctx, clientID, protocol.ResponseMessage{
		Type:      protocol.TypeResponse,
		Operation: "JOIN_ROOM",
		Result:    "SUCCESS",
		Extra:     request.RoomName,
	})

	joinedFrame := protocol.MustMarshal(protocol.JoinedRoomMessage{
		Type:     protocol.TypeJoinedRoom,
		RoomName: request.RoomName,
		Username: username,
	})

	// The spec says broadcast to users inside the room.
	// At this point, the user is inside the room, so they will receive it too.
	h.broadcastToRoomMembers(ctx, room, joinedFrame)
}

func (h *Hub) handleRoomUsers(
	ctx context.Context,
	requestingClientID ClientID,
	envelope protocol.Envelope,
) {
	request, err := protocol.DecodeRoomUsers(envelope)
	if err != nil {
		h.sendInvalidAndDisconnect(ctx, requestingClientID, "INVALID", "INVALID")
		return
	}

	if len(request.RoomName) == 0 || len(request.RoomName) > h.cfg.MaxRoomNameLength {
		h.sendInvalidAndDisconnect(ctx, requestingClientID, "INVALID", "INVALID")
		return
	}

	room, exists := h.rooms[request.RoomName]
	if !exists {
		h.sendResponse(ctx, requestingClientID, protocol.ResponseMessage{
			Type:      protocol.TypeResponse,
			Operation: "ROOM_USERS",
			Result:    "NO_SUCH_ROOM",
			Extra:     request.RoomName,
		})
		return
	}

	if _, isMember := room.members[requestingClientID]; !isMember {
		h.sendResponse(ctx, requestingClientID, protocol.ResponseMessage{
			Type:      protocol.TypeResponse,
			Operation: "ROOM_USERS",
			Result:    "NOT_JOINED",
			Extra:     request.RoomName,
		})
		return
	}

	roomUsersSnapshot := make(map[string]protocol.Status, len(room.members))
	for memberClientID := range room.members {
		memberUsername, isIdentified := h.clientUser[memberClientID]
		if !isIdentified {
			// Defensive: members should always be identified.
			continue
		}

		memberStatus, hasStatus := h.clientStatus[memberClientID]
		if !hasStatus {
			memberStatus = protocol.StatusActive
		}

		roomUsersSnapshot[memberUsername] = memberStatus
	}

	roomUserListFrame := protocol.MustMarshal(protocol.RoomUserListMessage{
		Type:     protocol.TypeRoomUserList,
		RoomName: request.RoomName,
		Users:    roomUsersSnapshot,
	})

	h.sendFrame(ctx, requestingClientID, roomUserListFrame)
}

func (h *Hub) handleRoomText(
	ctx context.Context,
	senderClientID ClientID,
	senderUsername string,
	envelope protocol.Envelope,
) {
	request, err := protocol.DecodeRoomText(envelope)
	if err != nil {
		h.sendInvalidAndDisconnect(ctx, senderClientID, "INVALID", "INVALID")
		return
	}

	if len(request.RoomName) == 0 || len(request.RoomName) > h.cfg.MaxRoomNameLength {
		h.sendInvalidAndDisconnect(ctx, senderClientID, "INVALID", "INVALID")
		return
	}

	room, exists := h.rooms[request.RoomName]
	if !exists {
		h.sendResponse(ctx, senderClientID, protocol.ResponseMessage{
			Type:      protocol.TypeResponse,
			Operation: "ROOM_TEXT",
			Result:    "NO_SUCH_ROOM",
			Extra:     request.RoomName,
		})
		return
	}

	if _, isMember := room.members[senderClientID]; !isMember {
		h.sendResponse(ctx, senderClientID, protocol.ResponseMessage{
			Type:      protocol.TypeResponse,
			Operation: "ROOM_TEXT",
			Result:    "NOT_JOINED",
			Extra:     request.RoomName,
		})
		return
	}

	roomTextFrame := protocol.MustMarshal(protocol.RoomTextFromMessage{
		Type:     protocol.TypeRoomTextFrom,
		RoomName: request.RoomName,
		Username: senderUsername,
		Text:     request.Text,
	})

	for memberClientID := range room.members {
		if memberClientID == senderClientID {
			continue
		}
		h.sendFrame(ctx, memberClientID, roomTextFrame)
	}
}

func (h *Hub) handleLeaveRoom(
	ctx context.Context,
	leavingClientID ClientID,
	leavingUsername string,
	envelope protocol.Envelope,
) {
	request, err := protocol.DecodeLeaveRoom(envelope)
	if err != nil {
		h.sendInvalidAndDisconnect(ctx, leavingClientID, "INVALID", "INVALID")
		return
	}

	if len(request.RoomName) == 0 || len(request.RoomName) > h.cfg.MaxRoomNameLength {
		h.sendInvalidAndDisconnect(ctx, leavingClientID, "INVALID", "INVALID")
		return
	}

	room, exists := h.rooms[request.RoomName]
	if !exists {
		h.sendResponse(ctx, leavingClientID, protocol.ResponseMessage{
			Type:      protocol.TypeResponse,
			Operation: "LEAVE_ROOM",
			Result:    "NO_SUCH_ROOM",
			Extra:     request.RoomName,
		})
		return
	}

	if _, isMember := room.members[leavingClientID]; !isMember {
		h.sendResponse(ctx, leavingClientID, protocol.ResponseMessage{
			Type:      protocol.TypeResponse,
			Operation: "LEAVE_ROOM",
			Result:    "NOT_JOINED",
			Extra:     request.RoomName,
		})
		return
	}

	// Remove membership.
	delete(room.members, leavingClientID)

	// Update reverse index.
	clientRoomSet, hasClientRooms := h.clientRooms[leavingClientID]
	if hasClientRooms {
		delete(clientRoomSet, request.RoomName)
		if len(clientRoomSet) == 0 {
			delete(h.clientRooms, leavingClientID)
		}
	}

	leftFrame := protocol.MustMarshal(protocol.LeftRoomMessage{
		Type:     protocol.TypeLeftRoom,
		RoomName: request.RoomName,
		Username: leavingUsername,
	})

	// Broadcast to remaining room members (sender excluded because they already left).
	for memberClientID := range room.members {
		h.sendFrame(ctx, memberClientID, leftFrame)
	}

	h.deleteRoomIfEmpty(request.RoomName, room)
}

func (h *Hub) handleDisconnect(
	ctx context.Context,
	clientID ClientID,
	username string,
	envelope protocol.Envelope,
) {
	_, err := protocol.DecodeDisconnect(envelope)
	if err != nil {
		h.sendInvalidAndDisconnect(ctx, clientID, "INVALID", "INVALID")
		return
	}

	h.forceDisconnect(ctx, clientID, fmt.Sprintf("client requested disconnect (user=%s)", username))
}

func (h *Hub) sendInvalidAndDisconnect(
	ctx context.Context,
	clientID ClientID,
	operation string,
	result string,
) {
	h.sendResponse(ctx, clientID, protocol.ResponseMessage{
		Type:      protocol.TypeResponse,
		Operation: operation,
		Result:    result,
	})

	h.forceDisconnect(
		ctx,
		clientID,
		fmt.Sprintf("protocol violation: operation=%s result=%s", operation, result),
	)
}

func (h *Hub) ensureClientRoomSet(clientID ClientID) map[string]struct{} {
	existingSet, exists := h.clientRooms[clientID]
	if exists {
		return existingSet
	}

	newSet := make(map[string]struct{})
	h.clientRooms[clientID] = newSet
	return newSet
}

func (h *Hub) isRoomMember(room *RoomState, clientID ClientID) bool {
	_, isMember := room.members[clientID]
	return isMember
}

func (h *Hub) broadcastToRoomMembers(
	ctx context.Context,
	room *RoomState,
	frame []byte,
) {
	for memberClientID := range room.members {
		h.sendFrame(ctx, memberClientID, frame)
	}
}

func (h *Hub) deleteRoomIfEmpty(roomName string, room *RoomState) {
	if len(room.members) != 0 {
		return
	}
	delete(h.rooms, roomName)
}

func (h *Hub) sendResponse(
	ctx context.Context,
	clientID ClientID,
	message protocol.ResponseMessage,
) {
	h.sendFrame(ctx, clientID, protocol.MustMarshal(message))
}

func (h *Hub) sendFrame(ctx context.Context, clientID ClientID, frame []byte) {
	writer, exists := h.clients[clientID]
	if !exists {
		return
	}

	if err := writer.Send(ctx, frame); err != nil {
		// Fail closed on outbound delivery issues to avoid leaking resources
		// and to keep hub state consistent.
		// Avoid blocking the hub if the unregister channel is full.
		h.requestUnregisterNonBlocking(clientID, fmt.Sprintf("send failed: %v", err))
	}
}

func (h *Hub) requestUnregisterNonBlocking(clientID ClientID, reason string) {
	unregisterEvent := UnregisterEvent{
		ClientID: clientID,
		Reason:   reason,
	}

	select {
	case h.unregister <- unregisterEvent:
		// Enqueued successfully.
	default:
		// If the queue is full, avoid blocking the hub.
		// Fail closed and disconnect immediately.
		h.forceDisconnect(context.Background(), clientID, reason)
	}
}

func (h *Hub) broadcastExcept(
	ctx context.Context,
	exceptClientID ClientID,
	frame []byte,
) {
	for clientID := range h.clients {
		if clientID == exceptClientID {
			continue
		}
		h.sendFrame(ctx, clientID, frame)
	}
}

func (h *Hub) leaveAllJoinedRoomsWithNotification(
	ctx context.Context,
	leavingClientID ClientID,
	leavingUsername string,
) {
	clientRoomSet, hasClientRooms := h.clientRooms[leavingClientID]
	if !hasClientRooms || len(clientRoomSet) == 0 {
		return
	}

	roomNames := make([]string, 0, len(clientRoomSet))
	for roomName := range clientRoomSet {
		roomNames = append(roomNames, roomName)
	}

	for _, roomName := range roomNames {
		room, exists := h.rooms[roomName]
		if !exists {
			continue
		}

		// Remove membership first, then notify remaining members.
		delete(room.members, leavingClientID)
		delete(room.invited, leavingClientID)

		leftRoomFrame := protocol.MustMarshal(protocol.LeftRoomMessage{
			Type:     protocol.TypeLeftRoom,
			RoomName: roomName,
			Username: leavingUsername,
		})

		for remainingMemberClientID := range room.members {
			h.sendFrame(ctx, remainingMemberClientID, leftRoomFrame)
		}

		h.deleteRoomIfEmpty(roomName, room)
	}

	delete(h.clientRooms, leavingClientID)
}

func (h *Hub) forceDisconnect(ctx context.Context, clientID ClientID, reason string) {
	writer, exists := h.clients[clientID]
	if !exists {
		return
	}

	username, hadUser := h.clientUser[clientID]

	// Notify others according to the protocol before removing state.
	if hadUser {
		h.leaveAllJoinedRoomsWithNotification(ctx, clientID, username)

		disconnectedFrame := protocol.MustMarshal(protocol.DisconnectedMessage{
			Type:     protocol.TypeDisconnected,
			Username: username,
		})

		h.broadcastExcept(ctx, clientID, disconnectedFrame)
	} else {
		// If the client never identified, it cannot be in rooms by protocol,
		// and DISCONNECTED cannot be formed (no username).
		delete(h.clientRooms, clientID)
	}

	delete(h.clients, clientID)
	delete(h.clientUser, clientID)
	delete(h.clientStatus, clientID)

	if hadUser {
		delete(h.usernameOwner, username)
	}

	if err := writer.Close(); err != nil {
		h.logger.Printf("client close error: %v", err)
	}

	h.logger.Printf("client disconnected: id=%s reason=%s", clientID, reason)
}

func (h *Hub) closeAll(reason string) {
	// Use background context to ensure best-effort cleanup even during shutdown cancellation.
	ctx := context.Background()

	for clientID := range h.clients {
		h.forceDisconnect(ctx, clientID, reason)
	}
}

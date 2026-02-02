#ifndef CHAT_PROTOCOL_H
#define CHAT_PROTOCOL_H

#include <stdbool.h>
#include <stddef.h>

enum {
    CHAT_USERNAME_MAX_LEN = 8,
    CHAT_ROOMNAME_MAX_LEN = 16,
};

typedef enum {
    CHAT_STATUS_ACTIVE,
    CHAT_STATUS_AWAY,
    CHAT_STATUS_BUSY,
} chat_status_t;

typedef enum {
    CHAT_MSG_INVALID,

    /* Client to Server */
    CHAT_MSG_IDENTIFY,
    CHAT_MSG_STATUS,
    CHAT_MSG_USERS,
    CHAT_MSG_TEXT,
    CHAT_MSG_PUBLIC_TEXT,
    CHAT_MSG_NEW_ROOM,
    CHAT_MSG_INVITE,
    CHAT_MSG_JOIN_ROOM,
    CHAT_MSG_ROOM_USERS,
    CHAT_MSG_ROOM_TEXT,
    CHAT_MSG_LEAVE_ROOM,
    CHAT_MSG_DISCONNECT,

    /* Server to Client */
    CHAT_MSG_RESPONSE,
    CHAT_MSG_NEW_USER,
    CHAT_MSG_NEW_STATUS,
    CHAT_MSG_USER_LIST,
    CHAT_MSG_TEXT_FROM,
    CHAT_MSG_PUBLIC_TEXT_FROM,
    CHAT_MSG_INVITATION,
    CHAT_MSG_JOINED_ROOM,
    CHAT_MSG_ROOM_USER_LIST,
    CHAT_MSG_ROOM_TEXT_FROM,
    CHAT_MSG_LEFT_ROOM,
    CHAT_MSG_DISCONNECTED,
} chat_msg_type_t;

const char *chat_msg_type_to_string(chat_msg_type_t type);
chat_msg_type_t chat_msg_type_from_string(const char *type_string);

const char *chat_status_to_string(chat_status_t status);
bool chat_status_from_string(const char *status_string,
                             chat_status_t *out_status);

bool chat_username_is_valid(const char *username);
bool chat_roomname_is_valid(const char *roomname);

#endif

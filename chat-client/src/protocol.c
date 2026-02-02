#include "protocol.h"

#include <ctype.h>
#include <string.h>

typedef struct {
    chat_msg_type_t type;
    const char *type_string;
} chat_type_map_entry_t;

static const chat_type_map_entry_t chat_type_map[] = {
    {CHAT_MSG_IDENTIFY, "IDENTIFY"},
    {CHAT_MSG_STATUS, "STATUS"},
    {CHAT_MSG_USERS, "USERS"},
    {CHAT_MSG_TEXT, "TEXT"},
    {CHAT_MSG_PUBLIC_TEXT, "PUBLIC_TEXT"},
    {CHAT_MSG_NEW_ROOM, "NEW_ROOM"},
    {CHAT_MSG_INVITE, "INVITE"},
    {CHAT_MSG_JOIN_ROOM, "JOIN_ROOM"},
    {CHAT_MSG_ROOM_USERS, "ROOM_USERS"},
    {CHAT_MSG_ROOM_TEXT, "ROOM_TEXT"},
    {CHAT_MSG_LEAVE_ROOM, "LEAVE_ROOM"},
    {CHAT_MSG_DISCONNECT, "DISCONNECT"},

    {CHAT_MSG_RESPONSE, "RESPONSE"},
    {CHAT_MSG_NEW_USER, "NEW_USER"},
    {CHAT_MSG_NEW_STATUS, "NEW_STATUS"},
    {CHAT_MSG_USER_LIST, "USER_LIST"},
    {CHAT_MSG_TEXT_FROM, "TEXT_FROM"},
    {CHAT_MSG_PUBLIC_TEXT_FROM, "PUBLIC_TEXT_FROM"},
    {CHAT_MSG_INVITATION, "INVITATION"},
    {CHAT_MSG_JOINED_ROOM, "JOINED_ROOM"},
    {CHAT_MSG_ROOM_USER_LIST, "ROOM_USER_LIST"},
    {CHAT_MSG_ROOM_TEXT_FROM, "ROOM_TEXT_FROM"},
    {CHAT_MSG_LEFT_ROOM, "LEFT_ROOM"},
    {CHAT_MSG_DISCONNECTED, "DISCONNECTED"},
};

static bool chat_is_printable_ascii_no_space(unsigned char ch) {
    if (ch < 0x21 || ch > 0x7E) {
        return false;
    }
    if (isspace(ch)) {
        return false;
    }
    return true;
}

static bool chat_is_printable_ascii(unsigned char ch) {
    return (ch >= 0x20 && ch <= 0x7E);
}

const char *chat_msg_type_to_string(chat_msg_type_t type) {
    for (size_t i = 0; i < sizeof(chat_type_map) / sizeof(chat_type_map[0]);
         i++) {
        if (chat_type_map[i].type == type) {
            return chat_type_map[i].type_string;
        }
    }

    return "INVALID";
}

chat_msg_type_t chat_msg_type_from_string(const char *type_string) {
    if (!type_string) {
        return CHAT_MSG_INVALID;
    }

    for (size_t i = 0; i < sizeof(chat_type_map) / sizeof(chat_type_map[0]);
         i++) {
        if (strcmp(chat_type_map[i].type_string, type_string) == 0) {
            return chat_type_map[i].type;
        }
    }

    return CHAT_MSG_INVALID;
}

const char *chat_status_to_string(chat_status_t status) {
    switch (status) {
    case CHAT_STATUS_ACTIVE:
        return "ACTIVE";
    case CHAT_STATUS_AWAY:
        return "AWAY";
    case CHAT_STATUS_BUSY:
        return "BUSY";
    default:
        return "ACTIVE";
    }
}

bool chat_status_from_string(const char *status_string,
                             chat_status_t *out_status) {
    if (!status_string || !out_status) {
        return false;
    }

    if (strcmp(status_string, "ACTIVE") == 0) {
        *out_status = CHAT_STATUS_ACTIVE;
        return true;
    }
    if (strcmp(status_string, "AWAY") == 0) {
        *out_status = CHAT_STATUS_AWAY;
        return true;
    }
    if (strcmp(status_string, "BUSY") == 0) {
        *out_status = CHAT_STATUS_BUSY;
        return true;
    }

    return false;
}

bool chat_username_is_valid(const char *username) {
    if (!username) {
        return false;
    }

    size_t length = strlen(username);
    if (length <= 0 || length > CHAT_USERNAME_MAX_LEN) {
        return false;
    }

    for (size_t i = 0; i < length; i++) {
        unsigned char ch = (unsigned char)username[i];
        if (!chat_is_printable_ascii_no_space(ch)) {
            return false;
        }
    }

    return true;
}

bool chat_roomname_is_valid(const char *roomname) {
    if (!roomname) {
        return false;
    }

    size_t length = strlen(roomname);
    if (length <= 0 || length > CHAT_ROOMNAME_MAX_LEN) {
        return false;
    }

    for (size_t i = 0; i < length; i++) {
        unsigned char ch = (unsigned char)roomname[i];
        if (!chat_is_printable_ascii(ch)) {
            return false;
        }
    }

    return true;
}

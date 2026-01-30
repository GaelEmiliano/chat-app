/* src/json_message.c */
#include "json_message.h"

#include <string.h>

static json_t *chat_json_object_with_type(chat_msg_type_t type) {
    json_t *object = json_object();
    if (!object) {
        return NULL;
    }

    if (json_object_set_new(object, "type",
                            json_string(chat_msg_type_to_string(type))) != 0) {
        json_decref(object);
        return NULL;
    }

    return object;
}

json_t *chat_json_build_identify(const char *username) {
    json_t *object = chat_json_object_with_type(CHAT_MSG_IDENTIFY);
    if (!object) {
        return NULL;
    }

    if (json_object_set_new(object, "username", json_string(username)) != 0) {
        json_decref(object);
        return NULL;
    }

    return object;
}

json_t *chat_json_build_status(chat_status_t status) {
    json_t *object = chat_json_object_with_type(CHAT_MSG_STATUS);
    if (!object) {
        return NULL;
    }

    if (json_object_set_new(object, "status",
                            json_string(chat_status_to_string(status))) != 0) {
        json_decref(object);
        return NULL;
    }

    return object;
}

json_t *chat_json_build_users(void) {
    return chat_json_object_with_type(CHAT_MSG_USERS);
}

json_t *chat_json_build_text(const char *username, const char *text) {
    json_t *object = chat_json_object_with_type(CHAT_MSG_TEXT);
    if (!object) {
        return NULL;
    }

    if (json_object_set_new(object, "username", json_string(username)) != 0 ||
        json_object_set_new(object, "text", json_string(text)) != 0) {
        json_decref(object);
        return NULL;
    }

    return object;
}

json_t *chat_json_build_public_text(const char *text) {
    json_t *object = chat_json_object_with_type(CHAT_MSG_PUBLIC_TEXT);
    if (!object) {
        return NULL;
    }

    if (json_object_set_new(object, "text", json_string(text)) != 0) {
        json_decref(object);
        return NULL;
    }

    return object;
}

json_t *chat_json_build_new_room(const char *roomname) {
    json_t *object = chat_json_object_with_type(CHAT_MSG_NEW_ROOM);
    if (!object) {
        return NULL;
    }

    if (json_object_set_new(object, "roomname", json_string(roomname)) != 0) {
        json_decref(object);
        return NULL;
    }

    return object;
}

json_t *chat_json_build_invite(const char *roomname,
                               const char *const *usernames,
                               size_t username_count) {
    json_t *object = chat_json_object_with_type(CHAT_MSG_INVITE);
    if (!object) {
        return NULL;
    }

    json_t *array = json_array();
    if (!array) {
        json_decref(object);
        return NULL;
    }

    for (size_t i = 0; i < username_count; i++) {
        if (json_array_append_new(array, json_string(usernames[i])) != 0) {
            json_decref(array);
            json_decref(object);
            return NULL;
        }
    }

    if (json_object_set_new(object, "roomname", json_string(roomname)) != 0 ||
        json_object_set_new(object, "usernames", array) != 0) {
        json_decref(array);
        json_decref(object);
        return NULL;
    }

    return object;
}

json_t *chat_json_build_join_room(const char *roomname) {
    json_t *object = chat_json_object_with_type(CHAT_MSG_JOIN_ROOM);
    if (!object) {
        return NULL;
    }

    if (json_object_set_new(object, "roomname", json_string(roomname)) != 0) {
        json_decref(object);
        return NULL;
    }

    return object;
}

json_t *chat_json_build_room_users(const char *roomname) {
    json_t *object = chat_json_object_with_type(CHAT_MSG_ROOM_USERS);
    if (!object) {
        return NULL;
    }

    if (json_object_set_new(object, "roomname", json_string(roomname)) != 0) {
        json_decref(object);
        return NULL;
    }

    return object;
}

json_t *chat_json_build_room_text(const char *roomname, const char *text) {
    json_t *object = chat_json_object_with_type(CHAT_MSG_ROOM_TEXT);
    if (!object) {
        return NULL;
    }

    if (json_object_set_new(object, "roomname", json_string(roomname)) != 0 ||
        json_object_set_new(object, "text", json_string(text)) != 0) {
        json_decref(object);
        return NULL;
    }

    return object;
}

json_t *chat_json_build_leave_room(const char *roomname) {
    json_t *object = chat_json_object_with_type(CHAT_MSG_LEAVE_ROOM);
    if (!object) {
        return NULL;
    }

    if (json_object_set_new(object, "roomname", json_string(roomname)) != 0) {
        json_decref(object);
        return NULL;
    }

    return object;
}

json_t *chat_json_build_disconnect(void) {
    return chat_json_object_with_type(CHAT_MSG_DISCONNECT);
}

bool chat_json_extract_type(json_t *root, chat_msg_type_t *out_type) {
    if (!root || !out_type) {
        return false;
    }
    if (!json_is_object(root)) {
        return false;
    }

    json_t *type_value = json_object_get(root, "type");
    if (!json_is_string(type_value)) {
        return false;
    }

    *out_type = chat_msg_type_from_string(json_string_value(type_value));
    return (*out_type != CHAT_MSG_INVALID);
}

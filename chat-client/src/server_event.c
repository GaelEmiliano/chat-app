/* src/server_event.c */
#include "server_event.h"

#include <string.h>

static const char *json_get_string_or_null(json_t *object, const char *key) {
    if (!json_is_object(object) || !key) {
        return NULL;
    }

    json_t *value = json_object_get(object, key);
    if (!json_is_string(value)) {
        return NULL;
    }

    return json_string_value(value);
}

static void print_kv_string(FILE *out, const char *label, const char *value) {
    if (!out || !label) {
        return;
    }

    if (value) {
        fprintf(out, "%s: %s\n", label, value);
    } else {
        fprintf(out, "%s: <missing>\n", label);
    }
}

static void print_user_map(FILE *out, json_t *users_object) {
    if (!out) {
        return;
    }

    if (!json_is_object(users_object)) {
        fprintf(out, "users: <missing>\n");
        return;
    }

    const char *username_key = NULL;
    json_t *status_value = NULL;

    fprintf(out, "users:\n");
    json_object_foreach(users_object, username_key, status_value) {
        const char *status_string = json_is_string(status_value)
                                        ? json_string_value(status_value)
                                        : "<invalid>";
        fprintf(out, "  - %s: %s\n", username_key ? username_key : "<invalid>",
                status_string);
    }
}

static bool extract_type(json_t *root, chat_msg_type_t *out_type) {
    const char *type_string = json_get_string_or_null(root, "type");
    if (!type_string) {
        return false;
    }

    *out_type = chat_msg_type_from_string(type_string);
    return (*out_type != CHAT_MSG_INVALID);
}

bool chat_server_event_print(json_t *root, FILE *output_stream) {
    if (!root || !output_stream) {
        return false;
    }

    chat_msg_type_t type = CHAT_MSG_INVALID;
    if (!extract_type(root, &type)) {
        fprintf(output_stream,
                "server: invalid message (missing/unknown type)\n");
        return false;
    }

    switch (type) {
    case CHAT_MSG_NEW_USER: {
        fprintf(output_stream, "[NEW_USER]\n");
        print_kv_string(output_stream, "username",
                        json_get_string_or_null(root, "username"));
        break;
    }
    case CHAT_MSG_NEW_STATUS: {
        fprintf(output_stream, "[NEW_STATUS]\n");
        print_kv_string(output_stream, "username",
                        json_get_string_or_null(root, "username"));
        print_kv_string(output_stream, "status",
                        json_get_string_or_null(root, "status"));
        break;
    }
    case CHAT_MSG_TEXT_FROM: {
        fprintf(output_stream, "[TEXT_FROM]\n");
        print_kv_string(output_stream, "from",
                        json_get_string_or_null(root, "username"));
        print_kv_string(output_stream, "text",
                        json_get_string_or_null(root, "text"));
        break;
    }
    case CHAT_MSG_PUBLIC_TEXT_FROM: {
        fprintf(output_stream, "[PUBLIC_TEXT_FROM]\n");
        print_kv_string(output_stream, "from",
                        json_get_string_or_null(root, "username"));
        print_kv_string(output_stream, "text",
                        json_get_string_or_null(root, "text"));
        break;
    }
    case CHAT_MSG_INVITATION: {
        fprintf(output_stream, "[INVITATION]\n");
        print_kv_string(output_stream, "from",
                        json_get_string_or_null(root, "username"));
        print_kv_string(output_stream, "roomname",
                        json_get_string_or_null(root, "roomname"));
        break;
    }
    case CHAT_MSG_JOINED_ROOM: {
        fprintf(output_stream, "[JOINED_ROOM]\n");
        print_kv_string(output_stream, "roomname",
                        json_get_string_or_null(root, "roomname"));
        print_kv_string(output_stream, "username",
                        json_get_string_or_null(root, "username"));
        break;
    }
    case CHAT_MSG_LEFT_ROOM: {
        fprintf(output_stream, "[LEFT_ROOM]\n");
        print_kv_string(output_stream, "roomname",
                        json_get_string_or_null(root, "roomname"));
        print_kv_string(output_stream, "username",
                        json_get_string_or_null(root, "username"));
        break;
    }
    case CHAT_MSG_DISCONNECTED: {
        fprintf(output_stream, "[DISCONNECTED]\n");
        print_kv_string(output_stream, "username",
                        json_get_string_or_null(root, "username"));
        break;
    }
    case CHAT_MSG_USER_LIST: {
        fprintf(output_stream, "[USER_LIST]\n");
        json_t *users =
            json_is_object(root) ? json_object_get(root, "users") : NULL;
        print_user_map(output_stream, users);
        break;
    }
    case CHAT_MSG_ROOM_USER_LIST: {
        fprintf(output_stream, "[ROOM_USER_LIST]\n");
        print_kv_string(output_stream, "roomname",
                        json_get_string_or_null(root, "roomname"));
        json_t *users =
            json_is_object(root) ? json_object_get(root, "users") : NULL;
        print_user_map(output_stream, users);
        break;
    }
    case CHAT_MSG_ROOM_TEXT_FROM: {
        fprintf(output_stream, "[ROOM_TEXT_FROM]\n");
        print_kv_string(output_stream, "roomname",
                        json_get_string_or_null(root, "roomname"));
        print_kv_string(output_stream, "from",
                        json_get_string_or_null(root, "username"));
        print_kv_string(output_stream, "text",
                        json_get_string_or_null(root, "text"));
        break;
    }
    case CHAT_MSG_RESPONSE: {
        fprintf(output_stream, "[RESPONSE]\n");
        print_kv_string(output_stream, "operation",
                        json_get_string_or_null(root, "operation"));
        print_kv_string(output_stream, "result",
                        json_get_string_or_null(root, "result"));
        print_kv_string(output_stream, "extra",
                        json_get_string_or_null(root, "extra"));
        break;
    }
    default:
        fprintf(output_stream, "[%s]\n", chat_msg_type_to_string(type));
        fprintf(output_stream,
                "server: message type recognized but not explicitly printed\n");
        break;
    }

    fflush(output_stream);
    return true;
}

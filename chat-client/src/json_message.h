/* src/json_message.h */
#ifndef JSON_MESSAGE_H
#define JSON_MESSAGE_H

#include "protocol.h"

#include <jansson.h>
#include <stdbool.h>
#include <stddef.h>

/*
 * JSON message builders.
 * Each function returns a new json_t object with refcount = 1.
 * Caller owns the returned object and must json_decref() it.
 */

/* Client to Server */
json_t *chat_json_build_identify(const char *username);
json_t *chat_json_build_status(chat_status_t status);
json_t *chat_json_build_users(void);
json_t *chat_json_build_text(const char *username, const char *text);
json_t *chat_json_build_public_text(const char *text);
json_t *chat_json_build_new_room(const char *roomname);
json_t *chat_json_build_invite(const char *roomname,
                               const char *const *usernames,
                               size_t username_count);
json_t *chat_json_build_join_room(const char *roomname);
json_t *chat_json_build_room_users(const char *roomname);
json_t *chat_json_build_room_text(const char *roomname, const char *text);
json_t *chat_json_build_leave_room(const char *roomname);
json_t *chat_json_build_disconnect(void);

/*
 * Minimal JSON parsing helpers.
 * These do not interpret the full message, only validate structure.
 */
bool chat_json_extract_type(json_t *root, chat_msg_type_t *out_type);

#endif

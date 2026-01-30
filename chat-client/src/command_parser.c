/* src/command_parser.c */
#include "command_parser.h"
#include "json_message.h"
#include "protocol.h"

#include <ctype.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

typedef struct {
    char **items;
    size_t count;
    size_t capacity;
} chat_token_list_t;

static chat_parse_result_t chat_parse_result_ok(chat_parse_action_t action,
                                                json_t *json_message) {
    chat_parse_result_t result;
    memset(&result, 0, sizeof(result));
    result.action = action;
    result.error = CHAT_PARSE_OK;
    result.json_message = json_message;
    return result;
}

static chat_parse_result_t chat_parse_result_error(chat_parse_error_t error,
                                                   const char *message) {
    chat_parse_result_t result;
    memset(&result, 0, sizeof(result));
    result.action = CHAT_PARSE_ACTION_NONE;
    result.error = error;

    if (message) {
        snprintf(result.error_message, sizeof(result.error_message), "%s",
                 message);
    } else {
        result.error_message[0] = '\0';
    }

    result.json_message = NULL;
    return result;
}

static void chat_token_list_init(chat_token_list_t *tokens) {
    tokens->items = NULL;
    tokens->count = 0;
    tokens->capacity = 0;
}

static void chat_token_list_destroy(chat_token_list_t *tokens) {
    if (!tokens) {
        return;
    }

    for (size_t i = 0; i < tokens->count; i++) {
        free(tokens->items[i]);
    }
    free(tokens->items);

    tokens->items = NULL;
    tokens->count = 0;
    tokens->capacity = 0;
}

static bool chat_token_list_push(chat_token_list_t *tokens, char *token_owned) {
    if (!tokens || !token_owned) {
        return false;
    }

    if (tokens->count == tokens->capacity) {
        size_t new_capacity =
            (tokens->capacity == 0) ? 8u : tokens->capacity * 2u;
        if (new_capacity < tokens->count) {
            return false;
        }

        char **new_items =
            realloc(tokens->items, new_capacity * sizeof(*new_items));
        if (!new_items) {
            return false;
        }

        tokens->items = new_items;
        tokens->capacity = new_capacity;
    }

    tokens->items[tokens->count++] = token_owned;
    return true;
}

static const char *chat_skip_spaces(const char *cursor) {
    while (*cursor != '\0' && isspace((unsigned char)*cursor)) {
        cursor++;
    }
    return cursor;
}

static bool chat_append_char(char **buffer, size_t *length, size_t *capacity,
                             char ch) {
    if (*length + 1u >= *capacity) {
        size_t new_capacity = (*capacity == 0) ? 64u : (*capacity * 2u);
        if (new_capacity < *length + 2u) {
            return false;
        }

        char *new_buffer = realloc(*buffer, new_capacity);
        if (!new_buffer) {
            return false;
        }

        *buffer = new_buffer;
        *capacity = new_capacity;
    }

    (*buffer)[(*length)++] = ch;
    (*buffer)[*length] = '\0';
    return true;
}

static bool chat_consume_escape(const char **cursor, char *out_char) {
    const char *p = *cursor;
    if (*p != '\\') {
        return false;
    }
    p++;

    char escaped = *p;
    if (escaped == '\0') {
        return false;
    }

    switch (escaped) {
    case 'n':
        *out_char = '\n';
        break;
    case 't':
        *out_char = '\t';
        break;
    case '\\':
        *out_char = '\\';
        break;
    case '"':
        *out_char = '"';
        break;
    default:
        *out_char = escaped;
        break;
    }

    p++;
    *cursor = p;
    return true;
}

/*
 * Tokenizer:
 *  - Splits by whitespace outside quotes.
 *  - Supports "quoted strings".
 *  - Supports backslash escapes everywhere.
 */
static chat_parse_error_t chat_tokenize_line(const char *line,
                                             chat_token_list_t *out_tokens,
                                             char *error_message,
                                             size_t error_message_size) {
    const char *cursor = chat_skip_spaces(line);
    while (*cursor != '\0') {
        bool in_quotes = false;
        char *token_buffer = NULL;
        size_t token_length = 0;
        size_t token_capacity = 0;

        if (*cursor == '"') {
            in_quotes = true;
            cursor++;
        }

        while (1) {
            if (*cursor == '\0') {
                if (in_quotes) {
                    snprintf(error_message, error_message_size,
                             "unterminated quote");
                    free(token_buffer);
                    return CHAT_PARSE_ERR_SYNTAX;
                }
                break;
            }

            if (!in_quotes && isspace((unsigned char)*cursor)) {
                break;
            }

            if (*cursor == '\\') {
                char decoded = '\0';
                const char *before = cursor;
                if (!chat_consume_escape(&cursor, &decoded)) {
                    snprintf(error_message, error_message_size,
                             "invalid escape sequence");
                    free(token_buffer);
                    return CHAT_PARSE_ERR_SYNTAX;
                }
                (void)before;
                if (!chat_append_char(&token_buffer, &token_length,
                                      &token_capacity, decoded)) {
                    free(token_buffer);
                    return CHAT_PARSE_ERR_NO_MEMORY;
                }
                continue;
            }

            if (in_quotes && *cursor == '"') {
                cursor++;
                in_quotes = false;
                break;
            }

            if (!chat_append_char(&token_buffer, &token_length, &token_capacity,
                                  *cursor)) {
                free(token_buffer);
                return CHAT_PARSE_ERR_NO_MEMORY;
            }
            cursor++;
        }

        if (!token_buffer) {
            token_buffer = strdup("");
            if (!token_buffer) {
                return CHAT_PARSE_ERR_NO_MEMORY;
            }
        }

        if (!chat_token_list_push(out_tokens, token_buffer)) {
            free(token_buffer);
            return CHAT_PARSE_ERR_NO_MEMORY;
        }

        cursor = chat_skip_spaces(cursor);
    }

    return CHAT_PARSE_OK;
}

static chat_parse_result_t
chat_build_and_validate_identify(const chat_token_list_t *tokens) {
    if (tokens->count < 2) {
        return chat_parse_result_error(CHAT_PARSE_ERR_MISSING_ARGUMENT,
                                       "missing username");
    }
    const char *username = tokens->items[1];
    if (!chat_username_is_valid(username)) {
        return chat_parse_result_error(CHAT_PARSE_ERR_INVALID_ARGUMENT,
                                       "invalid username");
    }

    json_t *msg = chat_json_build_identify(username);
    if (!msg) {
        return chat_parse_result_error(CHAT_PARSE_ERR_NO_MEMORY,
                                       "out of memory");
    }
    return chat_parse_result_ok(CHAT_PARSE_ACTION_SEND_JSON, msg);
}

static chat_parse_result_t
chat_build_and_validate_status(const chat_token_list_t *tokens) {
    if (tokens->count < 2) {
        return chat_parse_result_error(CHAT_PARSE_ERR_MISSING_ARGUMENT,
                                       "missing status");
    }

    chat_status_t status;
    if (!chat_status_from_string(tokens->items[1], &status)) {
        return chat_parse_result_error(
            CHAT_PARSE_ERR_INVALID_ARGUMENT,
            "invalid status (expected ACTIVE/AWAY/BUSY)");
    }

    json_t *msg = chat_json_build_status(status);
    if (!msg) {
        return chat_parse_result_error(CHAT_PARSE_ERR_NO_MEMORY,
                                       "out of memory");
    }
    return chat_parse_result_ok(CHAT_PARSE_ACTION_SEND_JSON, msg);
}

static chat_parse_result_t chat_build_users(void) {
    json_t *msg = chat_json_build_users();
    if (!msg) {
        return chat_parse_result_error(CHAT_PARSE_ERR_NO_MEMORY,
                                       "out of memory");
    }
    return chat_parse_result_ok(CHAT_PARSE_ACTION_SEND_JSON, msg);
}

static chat_parse_result_t chat_build_msg(const chat_token_list_t *tokens) {
    if (tokens->count < 3) {
        return chat_parse_result_error(CHAT_PARSE_ERR_MISSING_ARGUMENT,
                                       "usage: /msg <username> <text>");
    }

    const char *username = tokens->items[1];
    if (!chat_username_is_valid(username)) {
        return chat_parse_result_error(CHAT_PARSE_ERR_INVALID_ARGUMENT,
                                       "invalid username");
    }

    const char *text = tokens->items[2];
    if (text[0] == '\0') {
        return chat_parse_result_error(CHAT_PARSE_ERR_INVALID_ARGUMENT,
                                       "text must not be empty");
    }

    json_t *msg = chat_json_build_text(username, text);
    if (!msg) {
        return chat_parse_result_error(CHAT_PARSE_ERR_NO_MEMORY,
                                       "out of memory");
    }
    return chat_parse_result_ok(CHAT_PARSE_ACTION_SEND_JSON, msg);
}

static chat_parse_result_t chat_build_all(const chat_token_list_t *tokens) {
    if (tokens->count < 2) {
        return chat_parse_result_error(CHAT_PARSE_ERR_MISSING_ARGUMENT,
                                       "usage: /all <text>");
    }

    const char *text = tokens->items[1];
    if (text[0] == '\0') {
        return chat_parse_result_error(CHAT_PARSE_ERR_INVALID_ARGUMENT,
                                       "text must not be empty");
    }

    json_t *msg = chat_json_build_public_text(text);
    if (!msg) {
        return chat_parse_result_error(CHAT_PARSE_ERR_NO_MEMORY,
                                       "out of memory");
    }
    return chat_parse_result_ok(CHAT_PARSE_ACTION_SEND_JSON, msg);
}

static chat_parse_result_t chat_build_newroom(const chat_token_list_t *tokens) {
    if (tokens->count < 2) {
        return chat_parse_result_error(CHAT_PARSE_ERR_MISSING_ARGUMENT,
                                       "usage: /newroom <roomname>");
    }

    const char *roomname = tokens->items[1];
    if (!chat_roomname_is_valid(roomname)) {
        return chat_parse_result_error(CHAT_PARSE_ERR_INVALID_ARGUMENT,
                                       "invalid room name");
    }

    json_t *msg = chat_json_build_new_room(roomname);
    if (!msg) {
        return chat_parse_result_error(CHAT_PARSE_ERR_NO_MEMORY,
                                       "out of memory");
    }
    return chat_parse_result_ok(CHAT_PARSE_ACTION_SEND_JSON, msg);
}

static chat_parse_result_t chat_build_join(const chat_token_list_t *tokens) {
    if (tokens->count < 2) {
        return chat_parse_result_error(CHAT_PARSE_ERR_MISSING_ARGUMENT,
                                       "usage: /join <roomname>");
    }

    const char *roomname = tokens->items[1];
    if (!chat_roomname_is_valid(roomname)) {
        return chat_parse_result_error(CHAT_PARSE_ERR_INVALID_ARGUMENT,
                                       "invalid room name");
    }

    json_t *msg = chat_json_build_join_room(roomname);
    if (!msg) {
        return chat_parse_result_error(CHAT_PARSE_ERR_NO_MEMORY,
                                       "out of memory");
    }
    return chat_parse_result_ok(CHAT_PARSE_ACTION_SEND_JSON, msg);
}

static chat_parse_result_t
chat_build_roomusers(const chat_token_list_t *tokens) {
    if (tokens->count < 2) {
        return chat_parse_result_error(CHAT_PARSE_ERR_MISSING_ARGUMENT,
                                       "usage: /roomusers <roomname>");
    }

    const char *roomname = tokens->items[1];
    if (!chat_roomname_is_valid(roomname)) {
        return chat_parse_result_error(CHAT_PARSE_ERR_INVALID_ARGUMENT,
                                       "invalid room name");
    }

    json_t *msg = chat_json_build_room_users(roomname);
    if (!msg) {
        return chat_parse_result_error(CHAT_PARSE_ERR_NO_MEMORY,
                                       "out of memory");
    }
    return chat_parse_result_ok(CHAT_PARSE_ACTION_SEND_JSON, msg);
}

static chat_parse_result_t chat_build_roommsg(const chat_token_list_t *tokens) {
    if (tokens->count < 3) {
        return chat_parse_result_error(CHAT_PARSE_ERR_MISSING_ARGUMENT,
                                       "usage: /roommsg <roomname> <text>");
    }

    const char *roomname = tokens->items[1];
    if (!chat_roomname_is_valid(roomname)) {
        return chat_parse_result_error(CHAT_PARSE_ERR_INVALID_ARGUMENT,
                                       "invalid room name");
    }

    const char *text = tokens->items[2];
    if (text[0] == '\0') {
        return chat_parse_result_error(CHAT_PARSE_ERR_INVALID_ARGUMENT,
                                       "text must not be empty");
    }

    json_t *msg = chat_json_build_room_text(roomname, text);
    if (!msg) {
        return chat_parse_result_error(CHAT_PARSE_ERR_NO_MEMORY,
                                       "out of memory");
    }
    return chat_parse_result_ok(CHAT_PARSE_ACTION_SEND_JSON, msg);
}

static chat_parse_result_t chat_build_leave(const chat_token_list_t *tokens) {
    if (tokens->count < 2) {
        return chat_parse_result_error(CHAT_PARSE_ERR_MISSING_ARGUMENT,
                                       "usage: /leave <roomname>");
    }

    const char *roomname = tokens->items[1];
    if (!chat_roomname_is_valid(roomname)) {
        return chat_parse_result_error(CHAT_PARSE_ERR_INVALID_ARGUMENT,
                                       "invalid room name");
    }

    json_t *msg = chat_json_build_leave_room(roomname);
    if (!msg) {
        return chat_parse_result_error(CHAT_PARSE_ERR_NO_MEMORY,
                                       "out of memory");
    }
    return chat_parse_result_ok(CHAT_PARSE_ACTION_SEND_JSON, msg);
}

static chat_parse_result_t chat_build_invite(const chat_token_list_t *tokens) {
    if (tokens->count < 3) {
        return chat_parse_result_error(
            CHAT_PARSE_ERR_MISSING_ARGUMENT,
            "usage: /invite <roomname> <user1> [user2 ...]");
    }

    const char *roomname = tokens->items[1];
    if (!chat_roomname_is_valid(roomname)) {
        return chat_parse_result_error(CHAT_PARSE_ERR_INVALID_ARGUMENT,
                                       "invalid room name");
    }

    size_t username_count = tokens->count - 2u;
    const char *const *usernames = (const char *const *)&tokens->items[2];

    for (size_t index = 0; index < username_count; index++) {
        if (!chat_username_is_valid(usernames[index])) {
            return chat_parse_result_error(CHAT_PARSE_ERR_INVALID_ARGUMENT,
                                           "invalid username in invite list");
        }
    }

    json_t *msg = chat_json_build_invite(roomname, usernames, username_count);
    if (!msg) {
        return chat_parse_result_error(CHAT_PARSE_ERR_NO_MEMORY,
                                       "out of memory");
    }
    return chat_parse_result_ok(CHAT_PARSE_ACTION_SEND_JSON, msg);
}

static chat_parse_result_t chat_build_disconnect(void) {
    json_t *msg = chat_json_build_disconnect();
    if (!msg) {
        return chat_parse_result_error(CHAT_PARSE_ERR_NO_MEMORY,
                                       "out of memory");
    }
    return chat_parse_result_ok(CHAT_PARSE_ACTION_SEND_JSON, msg);
}

chat_parse_result_t chat_command_parse_line(const char *line) {
    if (!line) {
        return chat_parse_result_error(CHAT_PARSE_ERR_EMPTY, "empty input");
    }

    const char *trimmed = chat_skip_spaces(line);
    if (*trimmed == '\0') {
        return chat_parse_result_error(CHAT_PARSE_ERR_EMPTY, "empty input");
    }

    /* Non-command input: public text */
    if (*trimmed != '/') {
        json_t *msg = chat_json_build_public_text(trimmed);
        if (!msg) {
            return chat_parse_result_error(CHAT_PARSE_ERR_NO_MEMORY,
                                           "out of memory");
        }
        return chat_parse_result_ok(CHAT_PARSE_ACTION_SEND_JSON, msg);
    }

    /* Tokenize after the leading '/' */
    const char *command_line = trimmed + 1;
    chat_token_list_t tokens;
    chat_token_list_init(&tokens);

    char tokenize_error[160];
    tokenize_error[0] = '\0';

    chat_parse_error_t tokenize_status = chat_tokenize_line(
        command_line, &tokens, tokenize_error, sizeof(tokenize_error));

    if (tokenize_status != CHAT_PARSE_OK) {
        chat_token_list_destroy(&tokens);
        return chat_parse_result_error(
            tokenize_status,
            tokenize_error[0] ? tokenize_error : "command parse error");
    }

    if (tokens.count == 0 || tokens.items[0][0] == '\0') {
        chat_token_list_destroy(&tokens);
        return chat_parse_result_error(CHAT_PARSE_ERR_EMPTY, "empty command");
    }

    const char *command = tokens.items[0];
    chat_parse_result_t result;

    if (strcmp(command, "quit") == 0) {
        result = chat_parse_result_ok(CHAT_PARSE_ACTION_QUIT, NULL);
    } else if (strcmp(command, "identify") == 0) {
        result = chat_build_and_validate_identify(&tokens);
    } else if (strcmp(command, "status") == 0) {
        result = chat_build_and_validate_status(&tokens);
    } else if (strcmp(command, "users") == 0) {
        result = chat_build_users();
    } else if (strcmp(command, "msg") == 0) {
        result = chat_build_msg(&tokens);
    } else if (strcmp(command, "all") == 0) {
        result = chat_build_all(&tokens);
    } else if (strcmp(command, "newroom") == 0) {
        result = chat_build_newroom(&tokens);
    } else if (strcmp(command, "invite") == 0) {
        result = chat_build_invite(&tokens);
    } else if (strcmp(command, "join") == 0) {
        result = chat_build_join(&tokens);
    } else if (strcmp(command, "roomusers") == 0) {
        result = chat_build_roomusers(&tokens);
    } else if (strcmp(command, "roommsg") == 0) {
        result = chat_build_roommsg(&tokens);
    } else if (strcmp(command, "leave") == 0) {
        result = chat_build_leave(&tokens);
    } else if (strcmp(command, "disconnect") == 0) {
        result = chat_build_disconnect();
    } else {
        result = chat_parse_result_error(CHAT_PARSE_ERR_UNKNOWN_COMMAND,
                                         "unknown command");
    }

    chat_token_list_destroy(&tokens);
    return result;
}

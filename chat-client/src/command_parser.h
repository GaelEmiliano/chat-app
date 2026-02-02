#ifndef CHAT_COMMAND_PARSER_H
#define CHAT_COMMAND_PARSER_H

#include <jansson.h>
#include <stdbool.h>

typedef enum {
    CHAT_PARSE_ACTION_NONE,
    CHAT_PARSE_ACTION_SEND_JSON,
    CHAT_PARSE_ACTION_QUIT,
} chat_parse_action_t;

typedef enum {
    CHAT_PARSE_OK,
    CHAT_PARSE_ERR_EMPTY,
    CHAT_PARSE_ERR_SYNTAX,
    CHAT_PARSE_ERR_UNKNOWN_COMMAND,
    CHAT_PARSE_ERR_MISSING_ARGUMENT,
    CHAT_PARSE_ERR_INVALID_ARGUMENT,
    CHAT_PARSE_ERR_NO_MEMORY,
} chat_parse_error_t;

typedef struct {
    chat_parse_action_t action;
    chat_parse_error_t error;
    char error_message[160];
    json_t *json_message; /* owned by caller when action is SEND_JSON */
} chat_parse_result_t;

/*
 * Parses a single input line (without trailing '\n').
 *
 * Supported:
 *  - /identify <username>
 *  - /status ACTIVE|AWAY|BUSY
 *  - /users
 *  - /msg <username> <text...>
 *  - /all <text...>
 *  - /newroom <roomname>
 *  - /invite <roomname> <user1> [user2 ...]
 *  - /join <roomname>
 *  - /roomusers <roomname>
 *  - /roommsg <roomname> <text...>
 *  - /leave <roomname>
 *  - /disconnect
 *  - /quit
 *
 * Quoting:
 *  - Double quotes allow spaces: "Room 1"
 *  - Backslash escapes inside and outside quotes: \" \\ \n \t
 *
 * If the line does not start with '/', it is treated as public text.
 */
chat_parse_result_t chat_command_parse_line(const char *line);

#endif

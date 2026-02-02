#ifndef CHAT_APP_H
#define CHAT_APP_H

#include <stdbool.h>

/*
 * Runs the chat client main loop:
 *  - reads commands from stdin
 *  - reads JSON messages from the server
 *  - uses '\n' framing
 *
 * Returns true on clean shutdown, false on fatal error.
 */
bool chat_app_run(const char *server_host, const char *server_port);

#endif

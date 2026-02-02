#ifndef CHAT_NET_H
#define CHAT_NET_H

#include <stdbool.h>
#include <stddef.h>

typedef struct {
    int socket_fd;
} chat_net_connection_t;

/*
 * Establishes a TCP connection to host:port.
 * Returns true on success. On failure, the connection is left invalid.
 */
bool chat_net_connect(chat_net_connection_t *connection, const char *host,
                      const char *port);

/* Closes the connection if open and resets the descriptor. */
void chat_net_close(chat_net_connection_t *connection);

/*
 * Writes exactly byte_count bytes to fd.
 * Returns false on any unrecoverable error.
 */
bool chat_net_write_all(int fd, const void *buffer, size_t byte_count);

/*
 * Sends a UTF-8 JSON message using '\n' framing.
 * The JSON string must not contain the trailing newline.
 */
bool chat_net_send_json_line(int fd, const char *json_utf8);

#endif

/* src/net.c */
#include "net.h"

#include <errno.h>
#include <netdb.h>
#include <string.h>
#include <sys/socket.h>
#include <sys/types.h>
#include <unistd.h>

bool chat_net_connect(chat_net_connection_t *connection, const char *host,
                      const char *port) {
    if (!connection || !host || !port) {
        return false;
    }

    struct addrinfo hints;
    memset(&hints, 0, sizeof(hints));
    hints.ai_family = AF_UNSPEC;
    hints.ai_socktype = SOCK_STREAM;

    struct addrinfo *result = NULL;
    if (getaddrinfo(host, port, &hints, &result) != 0) {
        return false;
    }

    int socket_fd = -1;

    for (struct addrinfo *entry = result; entry; entry = entry->ai_next) {
        socket_fd =
            socket(entry->ai_family, entry->ai_socktype, entry->ai_protocol);
        if (socket_fd < 0) {
            continue;
        }

        if (connect(socket_fd, entry->ai_addr, entry->ai_addrlen) == 0) {
            break;
        }

        close(socket_fd);
        socket_fd = -1;
    }

    freeaddrinfo(result);

    if (socket_fd < 0) {
        return false;
    }

    connection->socket_fd = socket_fd;
    return true;
}

void chat_net_close(chat_net_connection_t *connection) {
    if (!connection) {
        return;
    }

    if (connection->socket_fd >= 0) {
        close(connection->socket_fd);
    }

    connection->socket_fd = -1;
}

bool chat_net_write_all(int fd, const void *buffer, size_t byte_count) {
    const unsigned char *bytes = buffer;
    size_t total_written = 0;

    while (total_written < byte_count) {
        ssize_t written =
            write(fd, bytes + total_written, byte_count - total_written);
        if (written > 0) {
            total_written += (size_t)written;
            continue;
        }

        if (written < 0 && errno == EINTR) {
            continue;
        }
        return false;
    }

    return true;
}

bool chat_net_send_json_line(int fd, const char *json_utf8) {
    if (!json_utf8) {
        return false;
    }

    size_t length = strlen(json_utf8);

    if (!chat_net_write_all(fd, json_utf8, length)) {
        return false;
    }
    if (!chat_net_write_all(fd, "\n", 1u)) {
        return false;
    }

    return true;
}

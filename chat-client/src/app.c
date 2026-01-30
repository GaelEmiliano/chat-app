/* src/app.c */
#include "app.h"
#include "command_parser.h"
#include "line_buffer.h"
#include "net.h"
#include "server_event.h"

#include <errno.h>
#include <jansson.h>
#include <poll.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

#define BUFFER_LENGTH 4096

static bool write_stderr_line(const char *message) {
    if (!message) {
        return false;
    }
    return (fprintf(stderr, "%s\n", message) >= 0);
}

static bool handle_server_input(int server_fd,
                                chat_line_buffer_t *server_buffer) {
    unsigned char read_buffer[BUFFER_LENGTH];

    ssize_t bytes_read = read(server_fd, read_buffer, sizeof(read_buffer));
    if (bytes_read == 0) {
        write_stderr_line("server: connection closed");
        return false;
    }
    if (bytes_read < 0) {
        if (errno == EINTR) {
            return true;
        }
        write_stderr_line("error: failed to read from server");
        return false;
    }

    if (!chat_line_buffer_append(server_buffer, read_buffer,
                                 (size_t)bytes_read)) {
        write_stderr_line("error: out of memory while buffering server input");
        return false;
    }

    while (1) {
        char *json_line = chat_line_buffer_pop_line(server_buffer);
        if (!json_line) {
            break;
        }

        json_error_t parse_error;
        memset(&parse_error, 0, sizeof(parse_error));

        json_t *root = json_loads(json_line, 0, &parse_error);
        if (!root) {
            fprintf(stderr, "server: invalid json: %s (line %d)\n",
                    parse_error.text, parse_error.line);
            free(json_line);
            continue;
        }

        (void)chat_server_event_print(root, stdout);

        json_decref(root);
        free(json_line);
    }

    return true;
}

static bool handle_stdin_input(int server_fd, chat_line_buffer_t *stdin_buffer,
                               bool *out_should_quit) {
    unsigned char read_buffer[BUFFER_LENGTH];

    ssize_t bytes_read = read(STDIN_FILENO, read_buffer, sizeof(read_buffer));
    if (bytes_read == 0) {
        *out_should_quit = true;
        return true;
    }
    if (bytes_read < 0) {
        if (errno == EINTR) {
            return true;
        }
        write_stderr_line("error: failed to read from stdin");
        return false;
    }

    if (!chat_line_buffer_append(stdin_buffer, read_buffer,
                                 (size_t)bytes_read)) {
        write_stderr_line("error: out of memory while buffering stdin");
        return false;
    }

    while (1) {
        char *input_line = chat_line_buffer_pop_line(stdin_buffer);
        if (!input_line) {
            break;
        }

        chat_parse_result_t parse_result = chat_command_parse_line(input_line);
        free(input_line);

        if (parse_result.error != CHAT_PARSE_OK) {
            if (parse_result.error_message[0] != '\0') {
                fprintf(stderr, "input: %s\n", parse_result.error_message);
            } else {
                fprintf(stderr, "input: parse error\n");
            }
            if (parse_result.json_message) {
                json_decref(parse_result.json_message);
            }
            continue;
        }

        if (parse_result.action == CHAT_PARSE_ACTION_QUIT) {
            if (parse_result.json_message) {
                json_decref(parse_result.json_message);
            }
            *out_should_quit = true;
            continue;
        }

        if (parse_result.action == CHAT_PARSE_ACTION_SEND_JSON &&
            parse_result.json_message) {
            char *json_compact = json_dumps(parse_result.json_message,
                                            JSON_COMPACT | JSON_ENSURE_ASCII);
            json_decref(parse_result.json_message);
            parse_result.json_message = NULL;

            if (!json_compact) {
                write_stderr_line(
                    "error: failed to serialize json (out of memory)");
                continue;
            }

            bool sent_ok = chat_net_send_json_line(server_fd, json_compact);
            free(json_compact);

            if (!sent_ok) {
                write_stderr_line("error: failed to send message to server");
                return false;
            }
        } else if (parse_result.json_message) {
            json_decref(parse_result.json_message);
        }
    }

    return true;
}

bool chat_app_run(const char *server_host, const char *server_port) {
    if (!server_host || !server_port) {
        return false;
    }

    chat_net_connection_t connection;
    connection.socket_fd = -1;

    if (!chat_net_connect(&connection, server_host, server_port)) {
        fprintf(stderr, "error: failed to connect to %s:%s\n", server_host,
                server_port);
        return false;
    }

    chat_line_buffer_t server_buffer;
    chat_line_buffer_t stdin_buffer;
    chat_line_buffer_init(&server_buffer);
    chat_line_buffer_init(&stdin_buffer);

    bool should_quit = false;
    bool success = true;

    while (!should_quit) {
        struct pollfd poll_fds[2];
        memset(poll_fds, 0, sizeof(poll_fds));

        poll_fds[0].fd = connection.socket_fd;
        poll_fds[0].events = POLLIN;

        poll_fds[1].fd = STDIN_FILENO;
        poll_fds[1].events = POLLIN;

        int poll_result = poll(poll_fds, 2, -1);
        if (poll_result < 0) {
            if (errno == EINTR) {
                continue;
            }
            write_stderr_line("error: poll failed");
            success = false;
            break;
        }

        if (poll_fds[0].revents & (POLLERR | POLLHUP | POLLNVAL)) {
            write_stderr_line("server: connection closed");
            break;
        }

        if (poll_fds[0].revents & POLLIN) {
            if (!handle_server_input(connection.socket_fd, &server_buffer)) {
                break;
            }
        }

        if (poll_fds[1].revents & (POLLERR | POLLHUP | POLLNVAL)) {
            should_quit = true;
        }

        if (poll_fds[1].revents & POLLIN) {
            if (!handle_stdin_input(connection.socket_fd, &stdin_buffer,
                                    &should_quit)) {
                success = false;
                break;
            }
        }
    }

    chat_line_buffer_destroy(&stdin_buffer);
    chat_line_buffer_destroy(&server_buffer);
    chat_net_close(&connection);

    return success;
}

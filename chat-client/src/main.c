/* src/main.c */
#include "app.h"

#include <stdio.h>

static void print_usage(const char *program_name) {
    fprintf(stderr, "usage: %s <host> <port>\n",
            program_name ? program_name : "chat-client");
}

int main(int argc, char **argv) {
    if (argc != 3) {
        print_usage((argc > 0) ? argv[0] : NULL);
        return 2;
    }

    const char *server_host = argv[1];
    const char *server_port = argv[2];

    if (!chat_app_run(server_host, server_port)) {
        return 1;
    }

    return 0;
}

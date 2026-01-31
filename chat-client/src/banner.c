#include "banner.h"

#include <stdio.h>

void chat_banner_print(FILE *output_stream) {
    if (!output_stream) {
        return;
    }

    fprintf(output_stream,
            "\n"
            "  ██████╗██╗  ██╗ █████╗ ████████╗\n"
            " ██╔════╝██║  ██║██╔══██╗╚══██╔══╝\n"
            " ██║     ███████║███████║   ██║   \n"
            " ██║     ██╔══██║██╔══██║   ██║   \n"
            " ╚██████╗██║  ██║██║  ██║   ██║   \n"
            "  ╚═════╝╚═╝  ╚═╝╚═╝  ╚═╝   ╚═╝   \n"
            "\n"
            "      Simple TCP Chat Client\n"
            "  --------------------------------\n"
            "\n"
            "  Commands:\n"
            "    \x1b[91m/identify\x1b[0m <username>\n"
            "    /status ACTIVE|AWAY|BUSY\n"
            "    \x1b[92m/users\x1b[0m \n"
            "\n"
            "    \x1b[93m/msg\x1b[0m <user> <text>\n"
            "    \x1b[94m/all\x1b[0m <text>\n"
            "\n"
            "    \x1b[95m/newroom\x1b[0m <room>\n"
            "    \x1b[31m/invite\x1b[0m <room> <user> [user ...]\n"
            "    \x1b[32m/join\x1b[0m <room>\n"
            "    \x1b[33m/roomusers\x1b[0m <room>\n"
            "    \x1b[34m/roommsg\x1b[0m <room> <text>\n"
            "    \x1b[35m/leave\x1b[0m <room>\n"
            "\n"
            "    \x1b[36m/disconnect\x1b[0m \n"
            "    /quit \n"
            "\n"
            "  Notes:\n"
            "   - You must /identify before using other commands\n"
            "   - Usernames no longer than 8 characters and without blanks\n"
            "   - Room names no longer than 16 characters\n"
            "   - Room names with spaces must be quoted\n"
            "   - Message text never needs quotes\n"
            "\n");

    fflush(output_stream);
}

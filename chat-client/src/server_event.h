#ifndef CHAT_SERVER_EVENT_H
#define CHAT_SERVER_EVENT_H

#include "protocol.h"

#include <jansson.h>
#include <stdbool.h>
#include <stdio.h>

/*
 * Prints a human-friendly representation of a server message.
 * This function never exits the program and should not crash on malformed JSON.
 *
 * Returns true if the message type was recognized (even if fields were
 * missing), false if the message was not recognized as a valid protocol
 * message.
 */
bool chat_server_event_print(json_t *root, FILE *output_stream);

#endif

/* src/line_buffer.c */
#include "line_buffer.h"

#include <stdlib.h>
#include <string.h>

void chat_line_buffer_init(chat_line_buffer_t *buffer) {
    if (!buffer) {
        return;
    }

    buffer->data = NULL;
    buffer->length = 0;
    buffer->capacity = 0;
}

void chat_line_buffer_destroy(chat_line_buffer_t *buffer) {
    if (!buffer) {
        return;
    }

    free(buffer->data);
    buffer->data = NULL;
    buffer->length = 0;
    buffer->capacity = 0;
}

static bool chat_line_buffer_reserve(chat_line_buffer_t *buffer,
                                     size_t required_capacity) {
    if (required_capacity <= buffer->capacity) {
        return true;
    }

    size_t new_capacity = (buffer->capacity == 0) ? 4096u : buffer->capacity;

    while (new_capacity < required_capacity) {
        if (new_capacity > (size_t)-1 / 2u) {
            return false;
        }
        new_capacity *= 2u;
    }

    char *new_data = realloc(buffer->data, new_capacity);
    if (!new_data) {
        return false;
    }

    buffer->data = new_data;
    buffer->capacity = new_capacity;
    return true;
}

bool chat_line_buffer_append(chat_line_buffer_t *buffer, const void *bytes,
                             size_t byte_count) {
    if (!buffer) {
        return false;
    }
    if (byte_count == 0) {
        return true;
    }
    if (!bytes) {
        return false;
    }

    if (buffer->length > (size_t)-1 - byte_count) {
        return false;
    }
    size_t required_capacity = buffer->length + byte_count;

    if (!chat_line_buffer_reserve(buffer, required_capacity)) {
        return false;
    }

    memcpy(buffer->data + buffer->length, bytes, byte_count);
    buffer->length += byte_count;
    return true;
}

char *chat_line_buffer_pop_line(chat_line_buffer_t *buffer) {
    if (!buffer || buffer->length == 0) {
        return NULL;
    }

    void *newline_ptr = memchr(buffer->data, '\n', buffer->length);
    if (!newline_ptr) {
        return NULL;
    }

    size_t line_length = (size_t)((char *)newline_ptr - buffer->data);

    char *line = malloc(line_length + 1u);
    if (!line) {
        return NULL;
    }

    memcpy(line, buffer->data, line_length);
    line[line_length] = '\0';

    size_t bytes_after_newline = buffer->length - (line_length + 1u);
    if (bytes_after_newline > 0) {
        memmove(buffer->data, buffer->data + line_length + 1u,
                bytes_after_newline);
    }
    buffer->length = bytes_after_newline;

    return line;
}

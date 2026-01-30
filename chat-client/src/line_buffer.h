/* src/line_buffer.h */
#ifndef LINE_BUFFER_H
#define LINE_BUFFER_H

#include <stdbool.h>
#include <stddef.h>

typedef struct {
    char *data;
    size_t length;
    size_t capacity;
} chat_line_buffer_t;

void chat_line_buffer_init(chat_line_buffer_t *buffer);
void chat_line_buffer_destroy(chat_line_buffer_t *buffer);

/* Appends raw bytes to the internal buffer. Returns false on OOM/overflow. */
bool chat_line_buffer_append(chat_line_buffer_t *buffer, const void *bytes,
                             size_t byte_count);

/*
 * Extracts one complete line (terminated by '\n') from the buffer.
 * - Returns a heap-allocated NUL-terminated string without the trailing '\n'.
 * - Returns NULL if no full line is available or on allocation failure.
 */
char *chat_line_buffer_pop_line(chat_line_buffer_t *buffer);

#endif

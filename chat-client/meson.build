project('clat-client', 'c',
  version: '0.1.0',
  default_options: [
    'c_std=c17',
    'warning_level=3',
    'werror=true',
  ]
)

cc = meson.get_compiler('c')

jansson_dep = dependency('jansson', required: true)

add_project_arguments(
  [
    '-D_POSIX_C_SOURCE=200809L',
  ],
  language: 'c'
)

sources = files(
  'src/app.c',
  'src/banner.c',
  'src/command_parser.c',
  'src/json_message.c',
  'src/line_buffer.c',
  'src/main.c',
  'src/net.c',
  'src/protocol.c',
  'src/server_event.c',
)

executable('chat-client',
  sources,
  dependencies: [jansson_dep],
  install: true
)

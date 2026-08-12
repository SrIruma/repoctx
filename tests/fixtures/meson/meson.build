project('mock-meson', 'c',
  version : '0.1.0',
)

gtk = dependency('gtk+-3.0')
glib = dependency('glib-2.0', required : false)

executable('mock-app', 'main.c',
  dependencies : [gtk, glib],
)

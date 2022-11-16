((
  {
    isDebugEnabled,
  } = pipy.solve('config.js'),
  global = {}
) => (

  global.debug = (f) => isDebugEnabled && f(console.log),

  global.shuffle = (arg, out, sort) => (
    arg && (
      sort = a => (a.map(e => e).map(() => a.splice(Math.random() * a.length | 0, 1)[0])),
      out = Object.fromEntries(sort(sort(Object.entries(arg))))
    ),
    out || {}
  ),

  global

))()
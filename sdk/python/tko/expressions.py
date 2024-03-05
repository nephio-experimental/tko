import types, collections.abc, tko.package, tko.plugin


def wrap(o):
  if isinstance(o, collections.abc.Mapping):
    r = types.SimpleNamespace()
    for k, v in o.items():
      setattr(r, k, wrap(v))
    return r
  if isinstance(o, list):
    r = []
    for v in o:
      r.append(wrap(v))
    return oo
  return o


def evaluate_expression(expression, package):
  if expression.startswith('{{') and expression.endswith('}}'):
    expression = expression[2:-2].strip()

    GVK = tko.package.GVK
    Identifier = tko.package.Identifier

    def get(g, v, k, name=None):
      gvk = GVK(g, v, k)
      if name is None:
        return wrap(package.get_first(gvk))
      else:
        return wrap(package[Identifier(gvk, name)])

    r = eval(expression)

    tko.plugin.log(f'evaluating: {expression} -> {r}')

    return r

  return e

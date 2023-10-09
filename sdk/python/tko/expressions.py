import tko.resources


def evaluate_expression(e, resources):
  if e.startswith('${') and e.endswith('}'):
    e = e[2:-1].strip()
    identifier, path = e.split('>', 2)
    identifier = identifier.strip().split(' ')

    if len(identifier) == 4:
      group, version, kind, name = identifier
    else:
      group = ''
      version, kind, name = identifier

    id = tko.resources.GVK(group=group, version=version, kind=kind)
    if name == '?':
      value = resources.get_first(id)
    else:
      id = tko.resources.Identifier(gvk=id, name=name)
      value = resources.get(id)

    if value is None:
      raise Exception(f'resource not found: {id}')

    segments = path.strip().split('.')
    for segment in segments:
      value = value.get(segment, {})

    if value == {}:
      value = ''

    return value

  return e

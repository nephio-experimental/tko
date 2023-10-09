import collections.abc


def validate_value(name, value, complete, validation):
  if not is_type(value, validation):
    raise Exception(f'"{name}" must be {validation}, is {type(value)}')

  if isinstance(validation, collections.abc.Mapping):
    if 'schema' in validation:
      validate_schema(name, value, complete, validation['schema'])
    if 'function' in validation:
      f = validation['function']
      if not callable(f):
        raise Exception('"function" is not callable')
      if not validation['function'](value):
        raise Exception(f'"{name}" failed validation function: {value}')


def validate_schema(name, value, complete, schema):
  if not isinstance(schema, collections.abc.Mapping):
    raise Exception('"schema" is not a mapping')
  if not isinstance(value, collections.abc.Mapping):
    raise Exception('"schema" is only applicable to mappings')

  for key, element in value.items():
    if key not in schema:
      raise Exception(f'unsupported key in "{name}": {key}')

  for key, value_schema in schema.items():
    if key in value:
      validate_value(f'{name}.{key}', value[key], complete, value_schema)
    elif complete and is_required(value_schema):
      raise Exception(f'missing required key "{name}": {key}')


def is_type(value, validation):
  if isinstance(validation, type):
    # Supports derived types
    return isinstance(value, validation)
  elif isinstance(validation, str):
    return value_type.__name__ == validation
  elif isinstance(validation, collections.abc.Mapping):
    if 'type' in validation:
      return is_type(value, validation['type'])
  elif isinstance(validation, collections.abc.Iterable):
    for validation_element in validation:
      if is_type(value, validation_element):
        return True
    return False
  return True


def is_required(validation):
  return isinstance(validation, collections.abc.Mapping) and (validation.get('required', False))

import collections.abc, cbor2
from io import BytesIO
from ruamel.yaml import YAML


yaml = YAML(typ='safe')
yaml.default_flow_style = False


def encode_package(package, format):
  if format == 'cbor':
    if isinstance(package, bytes):
      # CBOR bytes
      return package
    else:
      if not isinstance(package, collections.abc.Iterable):
        raise Exception('package is not iterable')

      # TODO: this cannot be unmarshalled from Go code on the server!
      buffer = BytesIO()
      cbor2.dump(package, buffer)
      return buffer.getvalue()

  elif format == 'yaml':
    if isinstance(package, bytes):
      # YAML text as bytes
      return package
    elif isinstance(package, str):
      # YAML text
      return bytes(package, 'utf-8')
    else:
      if not isinstance(package, collections.abc.Iterable):
        raise Exception('package is not iterable')

      buffer = BytesIO()
      first = True
      for resource in package:
        if first:
          first = False
        else:
          buffer.write(b'---\n')
        yaml.dump(resource, buffer)

      return buffer.getvalue()

  else:
    raise Exception(f'unsupported package format: {format}')


def decode_package(package, format):
  if format == 'cbor':
    return cbor2.loads(package)

  elif format == 'yaml':
    return yaml.load(package)

  else:
    raise Exception(f'unsupported package format: {format}')

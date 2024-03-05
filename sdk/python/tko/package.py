import io, collections.abc
from ruamel.yaml import YAML


yaml=YAML(typ='safe')
yaml.default_flow_style = False


class Package:
  def __init__(self, package=None):
    if package is not None:
      if not isinstance(package, collections.abc.Iterable):
        raise Exception('package is not iterable')
      for resource in package:
        if not isinstance(resource, collections.abc.Mapping):
          raise Exception('resource is not a mapping')
      self.package = package
    else:
      self.package = []

  def __str__(self):
    f = io.StringIO()
    first = True
    for resource in self:
      if not first:
        f.write('---\n')
      if first:
        first = False
      yaml.dump(resource, f)
    return f.getvalue()

  def __iter__(self):
    for resource in self.package:
      yield resource

  def __getitem__(self, identifier):
    for resource in self:
      if Identifier(resource=resource) == identifier:
        return resource
    return None

  def append(self, resource):
    if not isinstance(resource, collections.abc.Mapping):
      raise Exception('resource is not a mapping')
    self.package.append(resource)

  def iter_all(self, gvk):
    for resource in self:
      if GVK(resource=resource) == gvk:
        yield resource

  def get_first(self, gvk):
    for resource in self:
      if GVK(resource=resource) == gvk:
        return resource
    return None


class GVK:
  def __init__(self, group=None, version=None, kind=None, resource=None):
    if resource is not None:
      if not isinstance(resource, collections.abc.Mapping):
        raise Exception('resource is not a mapping')
      api_version = resource.get('apiVersion', '')
      if not isinstance(api_version, str):
        raise Exception('"apiVersion" is not a string')
      self.group, self.version = api_version.split('/', 2) if '/' in api_version else ('', api_version)
      self.kind = resource.get('kind', '')
      if not isinstance(self.kind, str):
        raise Exception('"kind" is not a string')
    elif (not isinstance(group, str)) or (not isinstance(version, str)) or (not isinstance(kind, str)):
        raise Exception('"group", "version", and "kind" are not all strings')
    else:
      self.group = group
      self.version = version
      self.kind = kind

  def __str__(self):
    return f'{self.group}, {self.version}, {self.kind}'

  def __eq__(self, o):
    return (self.group, self.version, self.kind) == (o.group, o.version, o.kind)

  @property
  def api_version(self):
    return f'{self.group}{"/" if self.group != "" else ""}{self.version}'


class Identifier:
  def __init__(self, gvk=None, name=None, resource=None):
    if resource is not None:
      self.gvk = GVK(resource=resource)
      self.name = resource.get('metadata', {}).get('name')
    elif (not isinstance(gvk, GVK)) or (not isinstance(name, str)):
      raise Exception('malformed identifier')
    else:
      self.gvk = gvk
      self.name = name

  def __str__(self):
    return f'{self.gvk}, {self.name}'

  def __eq__(self, o):
    return (self.gvk, self.name) == (o.gvk, o.name)


def set_prepared(resource, prepared=True):
  if not isinstance(resource, collections.abc.Mapping):
    raise Exception('resource is not a mapping')
  resource['metadata'] = resource.get('metadata', {})
  metadata = resource['metadata']
  metadata['annotations'] = metadata.get('annotations', {})
  annotations = metadata['annotations']
  if prepared:
    annotations['nephio.org/prepared'] = 'true'
  else:
    del annotations['nephio.org/prepared']

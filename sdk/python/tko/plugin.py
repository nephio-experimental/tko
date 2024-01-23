import sys, traceback, copy, tko.resources
from ruamel.yaml import YAML


yaml=YAML(typ='safe')
yaml.default_flow_style = False

input = None
output = {'prepared': False, 'resources': [], 'error': ''}
log_file = None


def get_output_resources():
  global output
  return tko.resources.Resources(output.get('resources', []))


def get_target_resource_identifier():
  global input
  target_resource_identifier = input.get('targetResourceIdentifier', {})
  group = target_resource_identifier.get('group', '')
  version = target_resource_identifier.get('version', '')
  kind = target_resource_identifier.get('kind', '')
  name = target_resource_identifier.get('name', '')
  gvk = tko.resources.GVK(group=group, version=version, kind=kind)
  return tko.resources.Identifier(gvk=gvk, name=name)


def get_target_resource():
  return get_output_resources()[get_target_resource_identifier()]


def get_deployments():
  global input
  deployments = input.get('deployments', {})
  for deployment in deployments.values():
    yield tko.Resources(deployment)


def get_grpc_host():
  global input
  grpc_ = input.get('grpc', {})
  protocol = grpc_.get('protocol', '')
  address = grpc_.get('address', '')
  port = grpc_.get('port', 0)
  if ':' in address:
    return f'[{address}]:{port}' # ipv6
  else:
    return f'{address}:{port}' # ipv4


def log(message):
  global log_file
  if log_file:
    log_file.write(message+'\n')


def open_log_file():
  global input, output, log_file
  log_file_ = input.get('logFile', '')
  if log_file_ != '':
    log_file = open(log_file_, 'w', buffering=1)


def validate(f):
  global input, output, log_file
  try:
    input = yaml.load(sys.stdin)
    open_log_file()
    output['resources'] = input.get('resources', [])
    complete = input.get('complete', True)
    f(complete)
  except:
    output['error'] = traceback.format_exc()

  del output['resources']
  yaml.dump(output, sys.stdout)


def prepare(f):
  global input, output, log_file
  try:
    input = yaml.load(sys.stdin)
    open_log_file()
    output['resources'] = copy.deepcopy(input.get('deploymentResources', []))
    if f():
      output['prepared'] = True
  except:
    output['resources'] = []
    output['error'] = traceback.format_exc()

  yaml.dump(output, sys.stdout)


def schedule(f):
  global input, output, log_file
  try:
    input = yaml.load(sys.stdin)
    open_log_file()
    output['resources'] = input.get('siteResources', [])
    f()
  except:
    output['error'] = traceback.format_exc()

  del output['resources']
  yaml.dump(output, sys.stdout)

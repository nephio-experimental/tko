import sys, traceback, copy, subprocess, os, collections.abc, tko.package, tko.encoding


input = None
output = {'package': [], 'error': ''}
log_file = None


def get_output_package():
  global output
  return tko.package.Package(output.get('package', []))


def get_target_resource_identifier():
  global input
  target_resource_identifier = input.get('targetResourceIdentifier', {})
  group = target_resource_identifier.get('group', '')
  version = target_resource_identifier.get('version', '')
  kind = target_resource_identifier.get('kind', '')
  name = target_resource_identifier.get('name', '')
  gvk = tko.package.GVK(group=group, version=version, kind=kind)
  return tko.package.Identifier(gvk=gvk, name=name)


def get_target_resource():
  return get_output_package()[get_target_resource_identifier()]


def get_deployments():
  global input
  deployments = input.get('deployments', {})
  for deployment in deployments.values():
    yield tko.package.Package(deployment)


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


def execute(args, env=None, input=None):
  if not isinstance(args, collections.abc.Sequence):
    raise Exception('args is not a sequence')
  log(f'executing: {" ".join(args)}')
  if env is not None:
      if not isinstance(env, collections.abc.Mapping):
        raise Exception('env is not a mapping')
      env_ = os.environ.copy()
      env_.update(env)
      env = env_
  if input is not None:
    input = input.encode()

  complete = subprocess.run(args, env=env, input=input, capture_output=True)
  if complete.returncode != 0:
    raise Exception(complete.stderr.decode())
  return complete.stdout.decode()


def log(message):
  global log_file
  if log_file:
    log_file.write(str(message)+'\n')


def open_log_file():
  global input, output, log_file
  log_file_ = input.get('logFile', '')
  if log_file_ != '':
    log_file = open(log_file_, 'w', buffering=1)


def validate(f):
  global input, output, log_file
  try:
    input = tko.encoding.yaml.load(sys.stdin)
    open_log_file()
    output['package'] = input.get('package', [])
    complete = input.get('complete', True)
    f(complete)
  except:
    output['error'] = traceback.format_exc()

  del output['package']
  tko.encoding.yaml.dump(output, sys.stdout)


def prepare(f):
  global input, output, log_file
  try:
    input = tko.encoding.yaml.load(sys.stdin)
    open_log_file()
    output['prepared'] = False
    output['package'] = copy.deepcopy(input.get('deploymentPackage', []))
    if f():
      output['prepared'] = True
  except:
    output['package'] = []
    output['error'] = traceback.format_exc()

  tko.encoding.yaml.dump(output, sys.stdout)


def schedule(f):
  global input, output, log_file
  try:
    input = tko.encoding.yaml.load(sys.stdin)
    open_log_file()
    output['package'] = input.get('sitePackage', [])
    f()
  except:
    output['error'] = traceback.format_exc()

  del output['package']
  tko.encoding.yaml.dump(output, sys.stdout)

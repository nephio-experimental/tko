import sys, traceback, copy, socket, subprocess, os, collections.abc, tko.package, tko.encoding


input = None
output = {'package': [], 'error': ''}
log_file = None
log_socket = None
log_socket_timeout = 5 # seconds


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


def execute(*args, env=None, input=None):
  env_ = ''.join(f'{k}={v} ' for k, v in env.items()) if env else ''
  args_ = " ".join(args)
  log(f'executing: {env_}{args_}')
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
  global log_file, log_socket
  if log_file:
    log_file.write(str(message)+'\n')
  elif log_socket:
    log_socket.sendall((str(message)+'\n').encode())


def open_log():
  global input, output, log_file, log_socket

  log_file_ = input.get('logFile', '')
  if log_file_ != '':
    try:
      log_file = open(log_file_, 'w', buffering=1)
    except Exception as e:
      raise Exception(f'cannot open life file {log_file_}: {e}')
    return

  log_address_port = input.get('logAddressPort', '')
  address, port = split_ip_address_port(log_address_port)
  if address is not None:
    try:
      if ':' in address:
        # IPv6
        family = socket.AF_INET6
      else:
        # IPv4
        family = socket.AF_INET
      # See: https://stackoverflow.com/a/4030559
      #      https://docs.python.org/3/library/socket.html#socket.getaddrinfo
      socket_address = socket.getaddrinfo(address, port, family, socket.SOCK_STREAM)[0][4]
      log_socket = socket.socket(family, socket.SOCK_STREAM)
      log_socket.settimeout(log_socket_timeout)
      log_socket.connect(socket_address)
    except Exception as e:
      raise Exception(f'cannot open log socket {log_address_port}: {e}')


def split_ip_address_port(address_port):
  split = address_port.rsplit(':', 1)
  if len(split) == 2:
    address = split[0]
    if address.startswith('['):
      # IPv6 (e.g. [fe80::aca9:89ff:fea8:f19b]:50055)
      address = address[1:-1]
    try:
      port = int(split[1])
    except:
      return (None, None)
    return (address, port)
  return (None, None)


def validate(f):
  global input, output, log_file
  try:
    input = tko.encoding.yaml.load(sys.stdin)
    open_log()
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
    open_log()
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
    open_log()
    output['package'] = input.get('sitePackage', [])
    f()
  except:
    output['error'] = traceback.format_exc()

  del output['package']
  tko.encoding.yaml.dump(output, sys.stdout)

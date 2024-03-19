import tko


def apply(package, kube_context=None):
  manifest = str(package)
  if manifest:
    execute('apply', '--filename', '-', input=manifest, kube_context=kube_context)


def execute(*args, input=None, kube_context=None):
  if kube_context is not None:
    args = list(args)
    args += ('--context', kube_context)
  return tko.execute('/usr/bin/kubectl', *args, input=input)

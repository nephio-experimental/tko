import tko


def execute(args, input=None, kube_context=None):
  if kube_context is not None:
    args = list(args)
    args.append(f'--context={kube_context}')
  return tko.execute(('/usr/bin/kubectl',) + tuple(args), input=input)


def apply(package, kube_context=None):
  manifest = str(package)
  if manifest:
    execute(('apply', '-f', '-'), input=manifest, kube_context=kube_context)

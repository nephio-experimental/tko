import subprocess, tko.plugin


def apply(package, context=None):
  manifest = str(package)
  if manifest:
    args = ['/usr/bin/kubectl', 'apply', '-f', '-']
    if context is not None:
      args.append(f'--context={context}')

    tko.plugin.execute(args, input=manifest.encode())

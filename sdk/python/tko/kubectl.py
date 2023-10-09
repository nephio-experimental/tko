import subprocess


env = {'PATH': '/usr/bin'}


def apply(resources, context=None):
  manifest = str(resources)
  if manifest:
    args = ['/usr/bin/kubectl', 'apply', '-f', '-']
    if context is not None:
      args.append(f'--context={context}')

    complete = subprocess.run(args, env=env, input=manifest.encode(), capture_output=True)
    if complete.returncode != 0:
      raise Exception(f'{complete.stderr.decode()}\n{manifest}')

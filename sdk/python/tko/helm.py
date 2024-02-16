import subprocess, json, tko.resources, tko.expressions


chart_gvk = tko.resources.GVK(group='workload.nephio.org', version='v1alpha1', kind='HelmChart')


def iter_charts(resources):
  return resources.iter_all(chart_gvk)


def get_current_deployments(context=None):
  env = {'PATH': '/usr/bin'}
  if context is not None:
    env['HELM_KUBECONTEXT'] = context

  complete = subprocess.run(('/usr/bin/helm', 'list', '--all-namespaces', '--deployed', '--short'), env=env, capture_output=True)
  if complete.returncode != 0:
    raise Exception(complete.stderr.decode())
  return complete.stdout.decode().rstrip('\n').split('\n')


def install(chart, resources, context=None):
  name = chart.get('metadata', {}).get('name', '')
  if name in get_current_deployments(context):
    return

  spec = chart.get('spec', {})
  repository = spec.get('repository', '')
  chart = spec.get('chart', '')
  if (repository == '') or (chart == ''):
    raise Exception('invalid HelmChart')

  args = ['/usr/bin/helm', 'install', '--replace', '--repo', repository, name, chart]

  env = {'PATH': '/usr/bin'}
  if context is not None:
    env['HELM_KUBECONTEXT'] = context

  parameters = spec.get('parameters', None)
  if parameters is not None:
    for key, value in parameters.items():
      value = tko.expressions.evaluate_expression(value, resources)
      args.append('--set-json')
      args.append(f'{key}={json.dumps(value)}')

  complete = subprocess.run(args, env=env, capture_output=True)
  if complete.returncode != 0:
    raise Exception(complete.stderr.decode())

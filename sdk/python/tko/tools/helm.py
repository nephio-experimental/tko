import json, tko


chart_gvk = tko.GVK(group='workload.nephio.org', version='v1alpha1', kind='HelmChart')


def execute(args, kube_context=None):
  env = {'HELM_NAMESPACE': 'default'}
  if kube_context is not None:
    env['HELM_KUBECONTEXT'] = kube_context

  return tko.execute(('/usr/bin/helm',) + tuple(args), env=env)


def iter_charts(package):
  return package.iter_all(chart_gvk)


def get_current_deployments(kube_context=None):
  return execute(('list', '--all-namespaces', '--deployed', '--short'), kube_context=kube_context).rstrip('\n').split('\n')


def install(chart, package, kube_context=None):
  name = chart.get('metadata', {}).get('name', '')
  if name in get_current_deployments(kube_context):
    return

  spec = chart.get('spec', {})
  chart = spec.get('chart', '')
  if chart == '':
    raise Exception('invalid HelmChart')

  args = ['install', '--replace']

  repository = spec.get('repository', '')
  if repository != '':
    args += ('--repo', repository)

  parameters = spec.get('parameters', None)
  if parameters is not None:
    for key, value in parameters.items():
      value = tko.evaluate_expression(value, package)
      args.append('--set-json')
      args.append(f'{key}={json.dumps(value)}')

  args += (name, chart)

  execute(args, kube_context=kube_context)

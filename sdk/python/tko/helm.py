import json, tko.plugin, tko.package, tko.expressions


chart_gvk = tko.package.GVK(group='workload.nephio.org', version='v1alpha1', kind='HelmChart')


def iter_charts(package):
  return package.iter_all(chart_gvk)


def get_current_deployments(context=None):
  args = ('/usr/bin/helm', 'list', '--all-namespaces', '--deployed', '--short')
  tko.plugin.log(f'executing: {" ".join(args)}')

  env = {}
  env['HELM_NAMESPACE'] = 'default'
  if context is not None:
    env['HELM_KUBECONTEXT'] = context

  return tko.plugin.execute(args, env=env).rstrip('\n').split('\n')


def install(chart, package, context=None):
  name = chart.get('metadata', {}).get('name', '')
  if name in get_current_deployments(context):
    return

  spec = chart.get('spec', {})
  chart = spec.get('chart', '')
  if chart == '':
    raise Exception('invalid HelmChart')

  args = ['/usr/bin/helm', 'install', '--replace' ]

  repository = spec.get('repository', '')
  if repository != '':
    args.extend(('--repo', repository))

  parameters = spec.get('parameters', None)
  if parameters is not None:
    for key, value in parameters.items():
      value = tko.expressions.evaluate_expression(value, package)
      args.append('--set-json')
      args.append(f'{key}={json.dumps(value)}')

  args.extend((name, chart))

  env = {}
  env['HELM_NAMESPACE'] = 'default'
  if context is not None:
    env['HELM_KUBECONTEXT'] = context

  tko.plugin.execute(args, env=env)

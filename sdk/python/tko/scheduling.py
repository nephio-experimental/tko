import tko.resources


def meta_schedule(resources):
  # TODO: this is too simplistic ;)
  # We should support per-user grouping and dependency-based ordering
  cluster_resources = tko.resources.Resources()
  namespaced_resources = tko.resources.Resources()

  for resource in resources:
    gvk = tko.resources.GVK(resource=resource)

    if gvk.group.endswith('nephio.org'):
      continue

    if is_namespaced(resource):
      namespaced_resources.append(resource)
    else:
      cluster_resources.append(resource)

  return cluster_resources, namespaced_resources


def is_namespaced(resource):
  return 'namespace' in resource.get('metadata', {})

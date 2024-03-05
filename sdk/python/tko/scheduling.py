import tko.package


def meta_schedule(package):
  # TODO: this is too simplistic ;)
  # We should support per-user grouping and dependency-based ordering
  cluster_package = tko.package.Package()
  namespaced_package = tko.package.Package()

  for resource in package:
    gvk = tko.package.GVK(resource=resource)

    if gvk.group.endswith('nephio.org'):
      continue

    if is_namespaced(resource):
      namespaced_package.append(resource)
    else:
      cluster_package.append(resource)

  return cluster_package, namespaced_package


def is_namespaced(resource):
  return 'namespace' in resource.get('metadata', {})

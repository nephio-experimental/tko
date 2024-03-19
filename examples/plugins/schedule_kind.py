#!/usr/bin/env python3

# Allow relative imports
import sys, pathlib
sys.path.append(str(pathlib.Path(__file__).parents[2] / 'sdk' / 'python'))

import tko
from tko.tools import kind, kubectl, helm


def schedule():
  cluster = tko.get_target_resource()
  if cluster is None:
    return
  if tko.GVK(resource=cluster) != kind.cluster_gvk:
    return

  cluster_name = kind.get_cluster_name(cluster)
  if cluster_name not in kind.get_current_cluster_names():
    kind.write_cluster_manifest(cluster)
    kind.create_cluster(cluster_name)

  kube_context = kind.get_kube_context(cluster_name)
  for deployment in tko.get_deployments():
    cluster_package, namespaced_package = tko.meta_schedule(deployment)
    kubectl.apply(cluster_package, kube_context=kube_context)
    kubectl.apply(namespaced_package, kube_context=kube_context)

    for chart in helm.iter_charts(deployment):
      helm.install(chart, deployment, kube_context=kube_context)


if __name__ == '__main__':
  tko.schedule(schedule)

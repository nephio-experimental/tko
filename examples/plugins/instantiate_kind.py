#!/usr/bin/env python3

# Allow relative imports
import sys, pathlib
sys.path.append(str(pathlib.Path(__file__).parents[2] / 'sdk' / 'python'))


import tko, tko.kind, tko.kubectl, tko.helm


def instantiate():
  # kind.x-k8s.io/v1alpha4, Cluster
  cluster = tko.get_target_resource()
  if cluster is not None:
    cluster_name = tko.kind.get_cluster_name(cluster)
    if cluster_name is not None:
      if cluster_name not in tko.kind.get_current_cluster_names():
        tko.kind.write_cluster_manifest(cluster)
        tko.log(f'creating Kind cluster: {cluster_name}')
        tko.kind.create_cluster()

    context = f'kind-{cluster_name}'
    for deployment in tko.get_deployments():
      cluster_resources, namespaced_resources = tko.meta_schedule(deployment)
      tko.kubectl.apply(cluster_resources, context=context)
      tko.kubectl.apply(namespaced_resources, context=context)

      for chart in tko.helm.iter_charts(deployment):
        #name = chart.get('metadata', {}).get('name', '')
        #tko.log(f'installing Helm chart for {name}')
        tko.helm.install(chart, deployment, context=context)


tko.instantiate(instantiate)

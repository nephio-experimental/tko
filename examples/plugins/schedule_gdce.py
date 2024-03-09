#!/usr/bin/env python3

# Allow relative imports
import sys, pathlib
sys.path.append(str(pathlib.Path(__file__).parents[2] / 'sdk' / 'python'))

import tko


edge_cluster_gvk = tko.GVK('gdce.google.com', 'v1alpha1', 'EdgeCluster')


def schedule():
  edge_cluster = tko.get_target_resource()
  if edge_cluster is None:
    return
  if tko.GVK(resource=edge_cluster) != edge_cluster_gvk:
    return

  #with tko.Client() as c:
  #  for s in c.list_sites():
  #    raise Exception(s)


if __name__ == '__main__':
  tko.schedule(schedule)

#!/usr/bin/env python3

# Allow relative imports
import sys, pathlib
sys.path.append(str(pathlib.Path(__file__).parents[2] / 'sdk' / 'python'))

import tko


def schedule():
  # gdce.google.com/v1alpha1, EdgeCluster
  cluster = tko.get_target_resource()
  if cluster is not None:
    pass
    #with tko.Client() as c:
    #  for s in c.list_sites():
    #    raise Exception(s)


tko.schedule(schedule)

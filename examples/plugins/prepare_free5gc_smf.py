#!/usr/bin/env python3

# Allow relative imports
import sys, pathlib
sys.path.append(str(pathlib.Path(__file__).parents[2] / 'sdk' / 'python'))

import tko


smf_gvk = tko.GVK('free5gc.plugin.nephio.org', 'v1alpha1', 'SMF')


def prepare():
  smf = tko.get_target_resource()
  if smf is None:
    return True
  if tko.GVK(resource=smf) != smf_gvk:
    return True

  smf['status'] = smf.get('status', {})
  smf['status']['test'] = 'hi'
  tko.set_prepared(smf)
  return True


if __name__ == '__main__':
  tko.prepare(prepare)

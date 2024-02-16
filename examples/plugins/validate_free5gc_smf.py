#!/usr/bin/env python3

# Allow relative imports
import sys, pathlib
sys.path.append(str(pathlib.Path(__file__).parents[2] / 'sdk' / 'python'))


import tko


smf_gvk = tko.GVK('free5gc.plugin.nephio.org', 'v1alpha1', 'SMF')


def validate(complete):
  # free5gc.plugin.nephio.org/v1alpha1, SMF
  smf = tko.get_target_resource()
  if smf is not None:
    if tko.GVK(resource=smf) != smf_gvk:
      return

    tko.validate_value('resource', smf, complete, {
      'type': dict,
      'schema': {
        'apiVersion': {'type': str, 'required': True},
        'kind': {'type': str, 'required': True},
        'metadata': dict,
        'spec': {
          'type': dict,
          'required': True,
          'schema': {
            'blahBlah': {
              'type': int,
              'required': True,
              'function': lambda x: x < 10
            },
            #'blahBlah': int,
            #'blahBlah': 'int',
            #'blahBlah': (str, int),
          }
        },
        'status': dict
      }
    })

    blah_blah = smf.get('spec', {}).get('blahBlah', None)
    if blah_blah is not None:
      if blah_blah > 10:
        raise Exception(f'"blahBlah" too big: {blah_blah}')


tko.validate(validate)

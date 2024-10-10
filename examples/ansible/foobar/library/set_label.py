from ansible.module_utils.basic import AnsibleModule
import tko


def run_module():
  module = AnsibleModule(argument_spec=dict(
    package=dict(type='raw'),
    label=dict(type='str'),
    value=dict(type='str'),
  ))

  package = module.params['package']

  for resource in package:
    metadata = resource['metadata'] = resource.get('metadata', {})
    labels = metadata['labels'] = metadata.get('labels', {})
    labels[module.params['label']] = module.params['value']

  module.exit_json(changed=False, package=package)


if __name__ == '__main__':
  run_module()

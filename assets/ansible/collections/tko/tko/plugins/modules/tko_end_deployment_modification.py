from ansible.module_utils.basic import AnsibleModule
import tko


def run_module():
  module = AnsibleModule(argument_spec=dict(
    host=dict(type='str', default='tko-data:50050'),

    modification_token=dict(type='str'),
    package=dict(type='raw'),
  ))

  with tko.Client(host=module.params['host']) as client:
    client.end_deployment_modification(module.params['modification_token'], package=module.params['package'])

  module.exit_json(changed=False)


if __name__ == '__main__':
  run_module()

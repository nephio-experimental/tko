from ansible.module_utils.basic import AnsibleModule
import tko


def run_module():
  module = AnsibleModule(argument_spec=dict(
    host=dict(type='str', default='tko-data:50050'),

    deployment_id=dict(type='str'),
  ))

  with tko.Client(host=module.params['host']) as client:
    modification_token, package = client.start_deployment_modification(module.params['deployment_id'])

  module.exit_json(changed=False, modification_token=modification_token, package=package)


if __name__ == '__main__':
  run_module()

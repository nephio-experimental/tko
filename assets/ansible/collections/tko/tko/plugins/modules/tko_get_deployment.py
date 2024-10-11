from ansible.module_utils.basic import AnsibleModule
import tko


def run_module():
  module = AnsibleModule(argument_spec=dict(
    host=dict(type='str', default='tko-data:50050'),

    deployment_id=dict(type='str'),
  ))

  with tko.Client(host=module.params['host']) as client:
    deployment = client.get_deployment(deployment_id=module.params['deployment_id'])
    deployment = deployment.to_ard()

  module.exit_json(changed=False, deployment=deployment)


if __name__ == '__main__':
  run_module()

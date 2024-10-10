from ansible.module_utils.basic import AnsibleModule
import tko


def run_module():
  module = AnsibleModule(argument_spec=dict(
    host=dict(type='str', default='tko-data:50050'),

    site_id=dict(type='str'),
    offset=dict(type='int', default=0),
    max_count=dict(type='int', default=1000),
  ))

  site_id = module.params['site_id']
  site_id_patterns = [site_id] if site_id else None

  with tko.Client(host=module.params['host']) as client:
    deployments = client.list_deployments(site_id_patterns=site_id_patterns, offset=module.params['offset'], max_count=module.params['max_count'])
    deployments = [deployment.to_ard() for deployment in deployments]

  module.exit_json(changed=False, deployments=deployments)


if __name__ == '__main__':
  run_module()

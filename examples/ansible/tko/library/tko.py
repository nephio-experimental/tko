from ansible.module_utils.basic import *
from google.protobuf.text_format import MessageToString
import tko


# https://docs.ansible.com/ansible/latest/dev_guide/developing_modules_general.html
# https://docs.ansible.com/ansible/latest/dev_guide/developing_module_utilities.html


def run_module():
    module = AnsibleModule(argument_spec=dict(
        offset=dict(type='int', default=0),
        max_count=dict(type='int', default=100),
    ))

    with tko.Client(host='tko-data:50050') as client:
        sites = client.list_sites(offset=module.params['offset'], max_count=module.params['max_count'])
        sites = [MessageToString(v) for v in sites]

    module.exit_json(changed=False, sites=sites)


if __name__ == '__main__':
    run_module()

from ansible.module_utils.basic import *
import tko


# https://docs.ansible.com/ansible/latest/dev_guide/developing_modules_general.html
# https://docs.ansible.com/ansible/latest/dev_guide/developing_module_utilities.html


def main():
    module = AnsibleModule(argument_spec={})

    client = tko.Client(host='tko-api:50050')
    value = {"hello": "world"}

    module.exit_json(changed=False, meta=value)


if __name__ == '__main__':
    main()

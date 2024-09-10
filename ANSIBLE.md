Ansible for TKO
===============

A PoC of using [Ansible](https://www.ansible.com/) playbooks to prepare and instantiate
deployments.

We rely on [Ansible AWX](https://www.ansible.com/awx/) to manage and run the playbooks.

Install
-------

Start with the local Kubernetes cluster environment (option 2 in the
[installation guide](INSTALL.md)). We will deploy AWX using the
[AWX operator](https://github.com/ansible/awx-operator), which is itself deployed
via a [Helm chart](https://github.com/ansible-community/awx-operator-helm):

    scripts/deploy-awx-kind

Note that it takes >5 minutes for AWX to come up.

Access
------

The web interface is at [http://localhost:30053](http://localhost:30053).
User: "admin", password: "tko".

The `awx-kind` script can be used to access the
[awx CLI](https://docs.ansible.com/automation-controller/latest/html/controllercli/),
e.g.:

    scripts/awx-kind projects list --conf.format=hum

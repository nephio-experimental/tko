AWX for TKO
===========

A PoC of using [Ansible](https://www.ansible.com/) playbooks to prepare and instantiate
deployments.

We rely on [Ansible AWX](https://www.ansible.com/awx/) (a.k.a. Ansible Tower) to manage
and run the playbooks.

Install
-------

Start with the local Kubernetes cluster environment (option 2 in the
[installation guide](INSTALL.md)). We will deploy AWX using the
[AWX operator](https://github.com/ansible/awx-operator), which is itself deployed
via a [Helm chart](https://github.com/ansible-community/awx-operator-helm):

    scripts/deploy-awx-kind

Note that it takes >5 minutes for AWX to come up (it spends a lot of time on data
migration, even if the database is empty.)

Access
------

The web interface is at [http://localhost:30053](http://localhost:30053).
User: "admin", password: "tko".

The `awx-kind` script can be used to access the
[awx CLI](https://docs.ansible.com/automation-controller/latest/html/controllercli/),
e.g.:

    scripts/awx-kind projects list

Test Scenario
-------------

To set up the test scenario run:

    scripts/test-awx-kind

(Note that if you get an error deleting the built in demo resources, just re-run the
script. It will eventually work. TODO: fix this!)

We are using a custom Ansible execution environment, built via
[this script](scripts/build-ansible-execution-environment). This minimal environment
contains the our [Python SDK](sdk/python/) and its dependencies.

The test scenario source is:

* [AWX project](examples/ansible/foobar/), registered as "Foobar".
* [tko.tko Ansible Galaxy collection](assets/ansible/collections/tko/tko/), providing
  modules that can be used by the playbooks.

The Foobar project contains:

* [`inventory-sites.py`](examples/ansible/foobar/inventory-sources/tko_sites.py): Inventory source for
  TKO sites. Sites can be selected using `TKO_METADATA_PATTERNS`. Using this source, two
  inventories are set up: "Chicago Sites" and "Bangalore Sites".
* [`provision-du.yaml`](examples/ansible/foobar/provision-cluster.yaml): Playbook that creates
  new Kind clusters for provisioned sites. It does this by accessing the "tko-runner" pod,
  where it provisions the clusters using Kind-on-Kind (using Docker-in-Docker). Registered as
  the "Provision DU" job template.
* [`prepare-du.yaml`](examples/ansible/foobar/prepare-du.yaml): Playbook that prepares
  deployments for scheduling. Uses the TKO collection to access the TKO data server. Registered
  as the "Prepare DU" job template.
* [`schedule-du.yaml`](examples/ansible/foobar/schedule-du.yaml): Playbook that schedules
  deployments on the Kind clusters. Registered as the "Schedule DU" job template.

We also create a workflow template named "Deploy DU". We create it via
[this playbook](examples/ansible/initialize-foobar.yaml), which strings together "Provision DU",
"Prepare DU", and "Schedule DU". (Due to current limitations in the `awx` CLI it is currently
impossible to create a workflow with it.)

Additionally, the project contains the
[`tko-plugin-schedule-kind.yaml`](examples/ansible/foobar/tko-plugin-schedule-kind.yaml)
playbook, which is registered as a scheduling plugin for TKO via the "ansible" executor. Note
that the reconciliation loop would normally keep running the playbook again and again, which is
undesirable for testing, so we limited it to only run once. To force the playbook to re-run,
delete its entry from AWX "jobs" (*not* "job templates"!).

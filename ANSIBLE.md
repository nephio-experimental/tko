Ansible for TKO
===============

[Ansible AWX](https://www.ansible.com/awx/).

Install
-------

Start with the local Kubernetes cluster environment (option 2 in the
[installation guide](INSTALL.md)).

    scripts/install-awx-kind

Note that it takes a few minutes for AWX to set up its database on PostgreSQL.

Access
------

The web interface is at [http://localhost:30053](http://localhost:30053).
User: "admin", password: "tko".

The `awk-kind` script can be used to access the
[awx CLI](https://docs.ansible.com/automation-controller/latest/html/controllercli/),
e.g.:

    scripts/awx-kind projects list -f human

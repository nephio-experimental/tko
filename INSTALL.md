TKO Installation Guide
======================

All general requirements can be installed by running:

    sudo scripts/install-system-dependencies
    scripts/install-python-dependencies

Note that Python will be using a virtual environment at `/tmp/tko-python-env`, so you will need to
reinstall the Python dependencies if you reboot.

### Fedora

    sudo scripts/install-system-dependencies-fedora

(also see [note](https://docs.fedoraproject.org/en-US/quick-docs/postgresql/) about editing
`/var/lib/pgsql/data/pg_hba.conf`, and make sure to also enable md5 for IPv6.)

### gLinux

    sudo scripts/install-system-dependencies-glinux

Or, these are the requirements if you prefer to install them manually:

* [Go](https://g3doc.corp.google.com/go/g3doc/codelabs/getting-started.md) (you should already have it in gLinux, but can install the latest version manually)
* [Docker](http://go/installdocker)
* [KIND](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)
* [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl-linux/)
* [Helm](https://helm.sh/docs/intro/install/)
* [kpt CLI](https://kpt.dev/installation/kpt-cli)
* [PostgreSQL](https://www.postgresql.org/)
* Python: ruamel.yaml

To setup our PostreSQL user:

    sudo scripts/setup-postgresql

Make sure Go-built binaries are in your path by adding this to your `~/.bashrc` file:

    export PATH="$HOME/go/bin:$PATH"

And then run this to use it now:

    . ~/.bashrc

Install our systemd services (in user mode) on top of the PostgreSQL backend:

    BACKEND=postgresql scripts/install-systemd-services

Finally, build tko, start the services, and deploy a few examples:

    scripts/test
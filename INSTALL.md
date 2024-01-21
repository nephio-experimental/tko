TKO Installation Guide
======================

Vagrant
-------

If you have [Vagrant](https://www.vagrantup.com/), we have a Vagrantfile ready for
a dev and test environment on top of a Fedora virtual machine. You'll need the
`vagrant-reload` plugin. To run:

    vagrant plugin install vagrant-reload
    cd tko
    vagrant up

It will take a few minutes.

The internal web server port will be mapped to your host at port 60051:
[http://localhost:60051/](http://localhost:60051/).

The virtual machine has the `tko` client. Example:

    vagrant ssh
    tko plugin list

The port is also mapped to the host at port 60050, so you could potentially run the client
there:

    tko plugin list --grpc-port 60050

If you want the virtual machine to to continuously sync file changes from the host (it's
one-way, only from the host to the virtual machine):

    vagrant rsync-auto

OS Requirements
---------------

### Fedora

    sudo scripts/install-system-dependencies-fedora

### gLinux

    sudo scripts/install-system-dependencies-glinux

Other Requirements
------------------

    sudo scripts/install-system-dependencies
    scripts/install-python-dependencies

Note that Python will be using a virtual environment at `~/tko-python-env`.

Or, these are the requirements if you prefer to install them manually:

* [Go](https://g3doc.corp.google.com/go/g3doc/codelabs/getting-started.md) (you should already have it in gLinux, but can install the latest version manually)
* [Docker](http://go/installdocker)
* [KIND](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)
* [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl-linux/)
* [Helm](https://helm.sh/docs/intro/install/)
* [kpt CLI](https://kpt.dev/installation/kpt-cli)
* [PostgreSQL](https://www.postgresql.org/)
* Python: ruamel.yaml

Setup
-----

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

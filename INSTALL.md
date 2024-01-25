TKO Installation Guide
======================

Vagrant Virtual Machine
-----------------------

We have a [Vagrantfile](https://www.vagrantup.com/) to create a dev and test environment
on top of a Fedora virtual machine. You'll need the `vagrant-reload` plugin. To run:

    cd tko
    vagrant plugin install vagrant-reload
    vagrant up

It will take a few minutes to install all dependencies. When done, it will reboot the
virtual machine and run the tests (see [testing](#testing) below).

The internal web server port is mapped to your host at port 60051, so you can access
the web UI at [http://localhost:60051/](http://localhost:60051/).

You can run `vagrant ssh` and then `cd /vagrant` to gain access to the environment.
We also provide a script to run commands on the virtual machine from the host. Examples:

    scripts/vagrant tko template list
    scripts/vagrant kubectl get pods --all-namespaces --context=kind-edge1
    scripts/vagrant scripts/test

If you have `tko` installed on the host, you can also run the client there. The API
gRPC port is mapped to the host at port 60050, so you need to point to it explicitly:

    tko plugin list --grpc-port=60050

Continue to [user guide](USAGE.md).

During development, if you want the virtual machine to continuously sync file changes
from the host (it's one-way, only from the host to the virtual machine at directory
`/vagrant`), run this in a separate terminal:

    vagrant rsync-auto

Native Installation
-------------------

### OS Requirements

For Fedora-family hosts:

    sudo scripts/install-system-dependencies-fedora

For Google gLinux hosts:

    sudo scripts/install-system-dependencies-glinux

### Other Requirements

    sudo scripts/install-system-dependencies
    scripts/install-python-dependencies

Note that Python will be using a virtual environment at `~/tko-python-env`.

If you're using the PostgreSQL backend, set up permissions:

    sudo scripts/setup-postgresql

These are the requirements if you prefer to install them manually:

* [Go](https://g3doc.corp.google.com/go/g3doc/codelabs/getting-started.md)
  (you should already have it in gLinux, but can still install the latest version manually)
* [Docker](https://docs.docker.com/get-docker/) ([instructions for Google gLinux](http://go/installdocker))
* [KIND](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)
* [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl-linux/)
* [Helm](https://helm.sh/docs/intro/install/)
* [kpt CLI](https://kpt.dev/installation/kpt-cli)
* [PostgreSQL](https://www.postgresql.org/)
* Python: [ruamel.yaml](https://pypi.org/project/ruamel.yaml/)

### Setup

Make sure Go-built binaries are in your path by adding this to your `~/.bashrc` file:

    export PATH="$HOME/go/bin:$PATH"

Also run that command locally to make it work in the current terminal.

Build TKO binaries:

    scripts/build

Install our systemd services (in user mode) on top of the PostgreSQL backend:

    BACKEND=postgresql BACKEND_CLEAN=true scripts/install-systemd-services

(By default it will install using the non-persistent memory backend, which is useful for
testing.)

Start the systemd services:

    scripts/start-service tko-api
    scripts/start-service tko-preparer
    scripts/start-service tko-meta-scheduler

Continue to [user guide](USAGE.md).

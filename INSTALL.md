TKO Installation Guide
======================

Option 1: Vagrant Virtual Machine
---------------------------------

We have a [Vagrantfile](https://www.vagrantup.com/) to create a dev and test environment
on top of a Fedora virtual machine (tested with both libvirt and VirtualBox providers).

To start it, run this in your local git clone directory:

    vagrant up && vagrant reload

It will take a few minutes to install all dependencies. *The reload is necessary for Docker
permissions to work.*

To run the test scenario:

    scripts/vagrant scripts/test

The `tko-data`'s web server port is mapped to your host at port 60051, so you can access
the web dashboard at [http://localhost:60051/](http://localhost:60051/).

To gain shell access to the dev environment run `vagrant ssh` and then `cd /vagrant`. Or,
to just run commands in the virtual machine use `scripts/vagrant`. Examples:

    scripts/vagrant tko template list
    scripts/vagrant tko dashboard
    scripts/vagrant kubectl get pods --all-namespaces --context=kind-edge1
    scripts/vagrant scripts/test

If you have `tko` installed on the host, you can also run the client there against the
virtual machine's `tko-data`'s gRPC port with this script:

    scripts/tko-vagrant dashboard

To follow logs from the host:

    scripts/vagrant scripts/log-service tko-data --follow
    scripts/vagrant scripts/log-service tko-preparer --follow
    scripts/vagrant scripts/log-service tko-meta-scheduler --follow

Also note that you can install the Kubernetes cluster (option 2 below) inside the Vagrant
virtual machine by running `scripts/test-kind`, combining both installation options. The
Kind's `tko-data`'s web server port is mapped to your host at port 60061.

During development, if you want the virtual machine to continuously sync file changes
from the host (it's one-way, only from the host to the virtual machine at directory
`/vagrant`), run this in a separate terminal:

    vagrant rsync-auto

To delete the virtual machine:

    vagrant destroy

Continue to the [user guide](USAGE.md), taking into account the scripts above.

Option 2: Local Kubernetes Cluster
----------------------------------

TKO can run in a Kubernetes cluster with a rich KRM aggregated API (in *addition* to the gRPC
API). We provide a quick setup on top of [Kind](https://kind.sigs.k8s.io/) using TKO container
images published on [Docker Hub](https://hub.docker.com/u/tliron). The setup includes a special
"runner" pod for executing plugins and deploying workload clusters, as well as PostgreSQL for
the TKO backend.

To create the Kind cluster locally and run the test scenario:

    scripts/test-kind

Note that you might get errors with pods related to too many open files. This is likely due to
your host's inotify limits being too low. See
[this](https://kind.sigs.k8s.io/docs/user/known-issues/#pod-errors-due-to-too-many-open-files)).

Also note that you can run `test-kind` inside the Vagrant virtual machine detailed above,
combining both installation options.

The `tko-data`'s web server port is mapped to your host at port 30051, so you can access
the web dashboard at [http://localhost:30051/](http://localhost:30051/).

If you have `tko` installed on the host, you can also run the client there against the
cluster's `tko-data`'s gRPC port with this script:

    scripts/tko-kind dashboard

`kubectl` access is provided with this script (it simply uses Kind's kube-config
context):

    scripts/kubectl-kind get tko
    scripts/kubectl-kind describe template/demo-002fhello-002dworld-003av1.0.0

See the [KRM API guide](KRM.md) for more information.

Plugins will be run in a pod named `tko-runner` and this also where the meta-scheduler's
Kind plugin will create clusters (Kind-in-Kind, using Docker-in-Docker, a.k.a. "dind").
We provide a script for accessing that environment:

    scripts/kind-runner kind get clusters
    scripts/kind-runner kubectl get pods --all-namespaces --context=kind-edge1

(Note that *this* use of `kubectl` is for the Kind-in-Kind workload clusters. For the
TKO "management" cluster use `scripts/kubectl-kind`.)

To follow logs from the host:

    scripts/log-service-kind tko-data --follow
    scripts/log-service-kind tko-preparer --follow
    scripts/log-service-kind tko-meta-scheduler --follow

To delete the Kind cluster:

    kind delete cluster --name=tko

Continue to the [user guide](USAGE.md), taking into account the scripts above.

Option 3: Native Installation
-----------------------------

### OS Requirements

For Fedora-family hosts:

    sudo scripts/install-system-dependencies-fedora

For Google gLinux hosts:

    sudo scripts/install-system-dependencies-glinux

You might have to reboot for Docker permissions to work.

### Other Requirements

    sudo scripts/install-system-dependencies
    scripts/install-python-env

Note that Python will be using a virtual environment at `~/tko-python-env`.

If you're using the PostgreSQL backend, set up permissions:

    sudo scripts/setup-postgresql

These are the requirements if you prefer to install them manually:

* [Go](https://g3doc.corp.google.com/go/g3doc/codelabs/getting-started.md)
  (you should already have it in gLinux, but can still install the latest version manually)
* [Docker](https://docs.docker.com/get-docker/) ([instructions for Google gLinux](http://go/installdocker))
* [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)
* [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl-linux/)
* [Helm](https://helm.sh/docs/intro/install/)
* [kpt CLI](https://kpt.dev/installation/kpt-cli)
* [PostgreSQL](https://www.postgresql.org/)
* Python: [ruamel.yaml](https://pypi.org/project/ruamel.yaml/)

### Setup

Make sure Go-built binaries are in your path by adding this to your `~/.bashrc` file:

    export PATH=$HOME/go/bin:$PATH

Also run that command locally to make it work in the current terminal.

Build TKO binaries:

    scripts/build

Install our systemd services (in user mode) on top of the PostgreSQL backend:

    BACKEND=postgresql BACKEND_CLEAN=true scripts/install-systemd-services

(By default it will install using the non-persistent memory backend, which is useful for
testing.)

Start the systemd services:

    scripts/start-services

To run the test scenario:

    scripts/test

We have a script to follow logs in individual tabs, supporting a few popular terminal
emulators (GNOME Terminal, Kitty, and Tilix):

    scripts/follow-logs

Or follow individual logs manually:

    scripts/log-service tko-data --follow
    scripts/log-service tko-preparer --follow
    scripts/log-service tko-meta-scheduler --follow

Continue to the [user guide](USAGE.md).
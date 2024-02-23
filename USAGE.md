TKO User Guide
==============

### Testing

We provide a test script that automatically builds and restarts the systemd services,
cleans up the PostgreSQL database (if used), and sets up all the examples. You can rerun
it any time:

    scripts/test

### TUI

Included is a rich TUI that even supports mouse clicks and scrolling:

    tko dashboard

It's fun to click on "Deployments", rerun `scripts/test`, and see the deployments being
created and prepared in real time!

### GUI

To access the web UI go to [http://localhost:50051](http://localhost:50051).

### CLI

The `tko` CLI is quite rich with CRUD commands for all entity types, as well as commands
for querying based on metadata and wildcards (all handled on the server). Quick example:

    tko deployment list --template-metadata=NetworkFunction.family=5G*

See the [test script](scripts/test) for more examples of CLI usage.

### Debugging

To follow the service logs (you can do this in separate terminal tabs):

    scripts/log-service tko-api --follow
    scripts/log-service tko-preparer --follow
    scripts/log-service tko-meta-scheduler --follow

To see the provisioned KIND clusters:

    kind get clusters

To access a provisioned Kind cluster use a context name where the cluster name is
prefixed with "kind-", e.g.:

    kubectl get pods --all-namespaces --context=kind-edge1

We provide a wrapper script for accessing the PostgreSQL CLI with our user. For
interactive mode:

    scripts/psql

For individual commands:

    scripts/psql -l

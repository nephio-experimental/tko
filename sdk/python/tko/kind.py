import os, pathlib, subprocess, copy, tko


kind_dir = pathlib.Path(__file__).parents[0] / 'kind'
kind_dir.mkdir(parents=True, exist_ok=True)
cluster_path = kind_dir / 'cluster.yaml'
env = {'PATH': '/usr/bin'}


def get_current_cluster_names():
  complete = subprocess.run(('/usr/bin/kind', 'get', 'clusters'), env=env, capture_output=True)
  if complete.returncode != 0:
    raise Exception(complete.stderr.decode())
  return complete.stdout.decode().rstrip('\n').split('\n')


def create_cluster():
  complete = subprocess.run(('/usr/bin/kind', 'create', 'cluster', '--config', cluster_path), env=env, capture_output=True)
  if complete.returncode != 0:
    raise Exception(complete.stderr.decode())


def write_cluster_manifest(cluster):
  with open(cluster_path, 'w') as f:
    if 'metadata' in cluster:
      cluster = copy.deepcopy(cluster)
      del cluster['metadata'] # Kind rejects metadata
    tko.yaml.dump(cluster, f)


def get_cluster_name(cluster):
  return cluster.get('name', '')

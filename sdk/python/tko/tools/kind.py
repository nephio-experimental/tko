import pathlib, copy, tko


cluster_gvk = tko.GVK(group='kind.x-k8s.io', version='v1alpha4', kind='Cluster')
kind_dir = pathlib.Path(__file__).parents[0] / 'kind'
kind_dir.mkdir(parents=True, exist_ok=True)
cluster_path = str(kind_dir / 'cluster.yaml')


def execute(args):
  return tko.execute(('/usr/bin/kind',) + tuple(args))


def get_current_cluster_names():
  return execute(('get', 'clusters')).rstrip('\n').split('\n')


def create_cluster():
  execute(('create', 'cluster', '--config', cluster_path))


def write_cluster_manifest(cluster):
  with open(cluster_path, 'w') as f:
    # Kind configuration resource doesn't support metadata, but kpt functions might add it
    if 'metadata' in cluster:
      cluster = copy.deepcopy(cluster)
      del cluster['metadata']
    tko.yaml.dump(cluster, f)


def get_cluster_name(cluster):
  return cluster.get('name', '')

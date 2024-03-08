import pathlib, copy, tko


kind_dir = pathlib.Path(__file__).parents[0] / 'kind'
kind_dir.mkdir(parents=True, exist_ok=True)
cluster_path = kind_dir / 'cluster.yaml'


def get_current_cluster_names():
  args = ('/usr/bin/kind', 'get', 'clusters')
  return tko.plugin.execute(args).rstrip('\n').split('\n')


def create_cluster():
  args = ('/usr/bin/kind', 'create', 'cluster', '--config', str(cluster_path))
  tko.plugin.execute(args)


def write_cluster_manifest(cluster):
  with open(cluster_path, 'w') as f:
    if 'metadata' in cluster:
      cluster = copy.deepcopy(cluster)
      del cluster['metadata'] # Kind rejects metadata
    tko.yaml.dump(cluster, f)


def get_cluster_name(cluster):
  return cluster.get('name', '')

import pathlib, copy, tko


cluster_gvk = tko.GVK(group='kind.x-k8s.io', version='v1alpha4', kind='Cluster')
kind_dir = pathlib.Path(__file__).parents[0] / 'kind'
kind_dir.mkdir(parents=True, exist_ok=True)


def get_cluster_name(cluster):
  return cluster.get('name', '')


def get_cluster_path(cluster_name):
  return kind_dir / f'{cluster_name}.yaml'


def get_kube_context(cluster_name):
  return f'kind-{cluster_name}'


def get_current_cluster_names():
  return execute('get', 'clusters').rstrip('\n').split('\n')


def create_cluster(cluster_name):
  tko.log(f'creating Kind cluster: {cluster_name}')
  execute('create', 'cluster', '--name', cluster_name, '--config', str(get_cluster_path(cluster_name)))


def write_cluster_manifest(cluster):
  cluster_path = get_cluster_path(get_cluster_name(cluster))
  with open(cluster_path, 'w') as f:
    # Kind configuration resource doesn't support metadata, but kpt functions might add it
    if 'metadata' in cluster:
      cluster = copy.deepcopy(cluster)
      del cluster['metadata']
    tko.yaml.dump(cluster, f)


def execute(*args):
  return tko.execute('/usr/bin/kind', *args)

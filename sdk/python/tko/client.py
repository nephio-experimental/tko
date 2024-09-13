import tko.plugin, tko.tko_pb2_grpc, tko.tko_pb2, grpc


MAX_INT32 = 1 << (32-1) - 1
DEFAULT_MAX_COUNT = 100
MAX_MAX_COUNT = MAX_INT32


class Client:
  def __init__(self, host=None):
    self.host = host if host is not None else tko.plugin.get_grpc_host()

  def __enter__(self):
    self.channel = grpc.insecure_channel(self.host)
    self.stub = tko.tko_pb2_grpc.APIStub(self.channel)
    return self

  def __exit__(self, *args):
    self.stub = None
    self.channel.close()
    self.channel = None

  def list_templates(self, template_id_patterns=[], metadata_patterns={}, offset=0, max_count=DEFAULT_MAX_COUNT):
    select_templates = tko.tko_pb2.SelectTemplates(templateIdPatterns=template_id_patterns, metadataPatterns=metadata_patterns)
    window = tko.tko_pb2.Window(offset=offset, maxCount=max_count)
    return self.stub.listTemplates(tko.tko_pb2.ListTemplates(select=select_templates, window=window))

  def list_sites(self, site_id_patterns=[], template_id_patterns=[], metadata_patterns={}, offset=0, max_count=DEFAULT_MAX_COUNT):
    select_sites = tko.tko_pb2.SelectSites(siteIdPatterns=site_id_patterns, templateIdPatterns=template_id_patterns, metadataPatterns=metadata_patterns)
    window = tko.tko_pb2.Window(offset=offset, maxCount=max_count)
    return self.stub.listSites(tko.tko_pb2.ListSites(select=select_sites, window=window))

  # "prepared" and "approved" can be True, False, or None (the default), which means either value.
  def list_deployments(self, parent_deployment_id=None, template_id_patterns=[], template_metadata_patterns={}, site_id_patterns=[], site_metadata_patterns={}, metadata_patterns={}, prepared=None, approved=None, offset=0, max_count=DEFAULT_MAX_COUNT):
    select_deployments = tko.tko_pb2.SelectDeployments(parentDeploymentId=parent_deployment_id, templateIdPatterns=template_id_patterns, templateMetadataPatterns=template_metadata_patterns, siteIdPatterns=site_id_patterns, siteMetadataPatterns=site_metadata_patterns, metadataPatterns=metadata_patterns, prepared=prepared, approved=approved)
    window = tko.tko_pb2.Window(offset=offset, maxCount=max_count)
    return self.stub.listDeployments(tko.tko_pb2.ListDeployments(select=select_deployments, window=window))

  # "trigger" is an optional dict with the keys "group", "version", and "kind".
  def list_plugins(self, type=None, name_patterns=[], executor=None, trigger=None, offset=0, max_count=DEFAULT_MAX_COUNT):
    select_plugins = tko.tko_pb2.SelectPlugins(type=type, namePatterns=name_patterns, executor=executor, trigger=trigger)
    window = tko.tko_pb2.Window(offset=offset, maxCount=max_count)
    return self.stub.listPlugins(tko.tko_pb2.ListPlugins(select=select_plugins, window=window))

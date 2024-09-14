import tko.plugin, tko.encoding, tko.tko_pb2_grpc, tko.tko_pb2, grpc


MAX_INT32 = 1 << (32-1) - 1
DEFAULT_MAX_COUNT = 100
MAX_MAX_COUNT = MAX_INT32
DEFAULT_PACKAGE_FORMAT = 'yaml'


def decode_package(self):
  return tko.encoding.decode_package(self.package, self.packageFormat)

tko.tko_pb2.Template.get_package = decode_package
tko.tko_pb2.Site.get_package = decode_package
tko.tko_pb2.Deployment.get_package = decode_package
tko.tko_pb2.StartDeploymentModificationResponse.get_package = decode_package


class Client:
  def __init__(self, host=None):
    self.host = host if host is not None else tko.plugin.get_grpc_host()

  def __enter__(self):
    self.channel = grpc.insecure_channel(self.host)
    self.stub = tko.tko_pb2_grpc.DataStub(self.channel)
    return self

  def __exit__(self, *args):
    self.stub = None
    self.channel.close()
    self.channel = None

  # Templates

  def register_template(self, template_id, metadata=None, package_format=DEFAULT_PACKAGE_FORMAT, package=None):
    if package is not None:
      package = tko.encoding.encode_package(package, format=package_format)
    r = self.stub.registerTemplate(tko.tko_pb2.Template(templateId=template_id, metadata=metadata, packageFormat=package_format, package=package))
    if not r.registered:
      raise Exception(r.notRegisteredReason)

  def get_template(self, template_id):
    return self.stub.getTemplate(tko.tko_pb2.GetTemplate(templateId=template_id))

  def delete_template(self, template_id):
    r = self.stub.deleteTemplate(tko.tko_pb2.TemplateID(templateId=template_id))
    if not r.deleted:
      raise Exception(r.notDeletedReason)

  def list_templates(self, template_id_patterns=None, metadata_patterns=None, offset=0, max_count=DEFAULT_MAX_COUNT):
    select_templates = tko.tko_pb2.SelectTemplates(templateIdPatterns=template_id_patterns, metadataPatterns=metadata_patterns)
    window = tko.tko_pb2.Window(offset=offset, maxCount=max_count)
    return self.stub.listTemplates(tko.tko_pb2.ListTemplates(select=select_templates, window=window))

  def purge_templates(self, template_id_patterns=None, metadata_patterns=None):
    r = self.stub.purgeTemplates(tko.tko_pb2.SelectTemplates(templateIdPatterns=template_id_patterns, metadataPatterns=metadata_patterns))
    if not r.deleted:
      raise Exception(r.notDeletedReason)

  # Sites

  def register_site(self, site_id, template_id=None, metadata=None, package_format=DEFAULT_PACKAGE_FORMAT, package=None):
    if package is not None:
      package = tko.encoding.encode_package(package, format=package_format)
    r = self.stub.registerSite(tko.tko_pb2.Site(siteId=site_id, templateId=template_id, metadata=metadata, packageFormat=package_format, package=package))
    if not r.registered:
      raise Exception(r.notRegisteredReason)

  def get_site(self, site_id):
    return self.stub.getSite(tko.tko_pb2.GetSite(siteId=site_id))

  def delete_site(self, site_id):
    r = self.stub.deleteSite(tko.tko_pb2.SiteID(siteId=site_id))
    if not r.deleted:
      raise Exception(r.notDeletedReason)

  def list_sites(self, site_id_patterns=None, template_id_patterns=None, metadata_patterns=None, offset=0, max_count=DEFAULT_MAX_COUNT):
    select_sites = tko.tko_pb2.SelectSites(siteIdPatterns=site_id_patterns, templateIdPatterns=template_id_patterns, metadataPatterns=metadata_patterns)
    window = tko.tko_pb2.Window(offset=offset, maxCount=max_count)
    return self.stub.listSites(tko.tko_pb2.ListSites(select=select_sites, window=window))

  def purge_sites(self, site_id_patterns=None, template_id_patterns=None, metadata_patterns=None):
    r = self.stub.purgeSites(tko.tko_pb2.SelectSites(siteIdPatterns=site_id_patterns, templateIdPatterns=template_id_patterns, metadataPatterns=metadata_patterns))
    if not r.deleted:
      raise Exception(r.notDeletedReason)

  # Deployments

  def create_deployment(self, template_id, parent_deployment_id=None, site_id=None, merge_metadata=None, prepared=False, approved=False, merge_package_format=DEFAULT_PACKAGE_FORMAT, merge_package=None):
    if merge_package is not None:
      merge_package = tko.encoding.encode_package(merge_package, format=merge_package_format)
    r = self.stub.createDeployment(tko.tko_pb2.CreateDeployment(templateId=template_id, parentDeploymentId=parent_deployment_id, siteId=site_id, mergeMetadata=merge_metadata, prepared=prepared, approved=approved, mergePackageFormat=merge_package_format, mergePackage=merge_package))
    if r.created:
      return r.deploymentId
    else:
      raise Exception(r.notCreatedReason)

  def get_deployment(self, deployment_id):
    return self.stub.getDeployment(tko.tko_pb2.GetDeployment(deploymentId=deployment_id))

  def delete_deployment(self, deployment_id):
    r = self.stub.deleteDeployment(tko.tko_pb2.DeploymentID(deploymentId=deployment_id))
    if not r.deleted:
      raise Exception(r.notDeletedReason)

  # "prepared" and "approved" can be True, False, or None (the default), which means either value.
  def list_deployments(self, parent_deployment_id=None, template_id_patterns=None, template_metadata_patterns=None, site_id_patterns=None, site_metadata_patterns=None, metadata_patterns=None, prepared=None, approved=None, offset=0, max_count=DEFAULT_MAX_COUNT):
    select_deployments = tko.tko_pb2.SelectDeployments(parentDeploymentId=parent_deployment_id, templateIdPatterns=template_id_patterns, templateMetadataPatterns=template_metadata_patterns, siteIdPatterns=site_id_patterns, siteMetadataPatterns=site_metadata_patterns, metadataPatterns=metadata_patterns, prepared=prepared, approved=approved)
    window = tko.tko_pb2.Window(offset=offset, maxCount=max_count)
    return self.stub.listDeployments(tko.tko_pb2.ListDeployments(select=select_deployments, window=window))

  # "prepared" and "approved" can be True, False, or None (the default), which means either value.
  def purge_deployments(self, parent_deployment_id=None, template_id_patterns=None, template_metadata_patterns=None, site_id_patterns=None, site_metadata_patterns=None, metadata_patterns=None, prepared=None, approved=None):
    r = self.stub.purgeDeployments(tko.tko_pb2.SelectDeployments(parentDeploymentId=parent_deployment_id, templateIdPatterns=template_id_patterns, templateMetadataPatterns=template_metadata_patterns, siteIdPatterns=site_id_patterns, siteMetadataPatterns=site_metadata_patterns, metadataPatterns=metadata_patterns, prepared=prepared, approved=approved))
    if not r.deleted:
      raise Exception(r.notDeletedReason)

  def start_deployment_modification(self, deployment_id):
    r = self.stub.startDeploymentModification(tko.tko_pb2.StartDeploymentModification(deploymentId=deployment_id))
    if r.started:
      return r.modificationToken, r
    else:
      raise Exception(r.notStartedReason)

  def end_deployment_modification(self, modification_token, package_format=DEFAULT_PACKAGE_FORMAT, package=None):
    if package is not None:
      package = tko.encoding.encode_package(package, format=package_format)
    r = self.stub.endDeploymentModification(tko.tko_pb2.EndDeploymentModification(modificationToken=modification_token, packageFormat=package_format, package=package))
    if not r.modified:
      raise Exception(r.notModifiedReason)

  def cancel_deployment_modification(self, modification_token):
    r = self.stub.cancelDeploymentModification(tko.tko_pb2.CancelDeploymentModification(modificationToken=modification_token))
    if not r.cancelled:
      raise Exception(r.notCancelledReason)

  # Plugins

  # "triggers" is an optional list of dicts with the keys "group", "version", and "kind".
  def register_plugin(self, type, name, executor, arguments=None, properties=None, triggers=None):
    r = self.stub.registerPlugin(tko.tko_pb2.Plugin(type=type, name=name, executor=executor, arguments=arguments, properties=properties, triggers=triggers))
    if not r.registered:
      raise Exception(r.notRegisteredReason)

  def get_plugin(self, type, name):
    return self.stub.getPlugin(tko.tko_pb2.PluginID(type=type, name=name))

  def delete_plugin(self, type, name):
    r = self.stub.deletePlugin(tko.tko_pb2.PluginID(type=type, name=name))
    if not r.deleted:
      raise Exception(r.notDeletedReason)

  # "trigger" is an optional dict with the keys "group", "version", and "kind".
  def list_plugins(self, type=None, name_patterns=None, executor=None, trigger=None, offset=0, max_count=DEFAULT_MAX_COUNT):
    select_plugins = tko.tko_pb2.SelectPlugins(type=type, namePatterns=name_patterns, executor=executor, trigger=trigger)
    window = tko.tko_pb2.Window(offset=offset, maxCount=max_count)
    return self.stub.listPlugins(tko.tko_pb2.ListPlugins(select=select_plugins, window=window))

  # "trigger" is an optional dict with the keys "group", "version", and "kind".
  def purge_plugins(self, type=None, name_patterns=None, executor=None, trigger=None):
    r = self.stub.purgePlugins(tko.tko_pb2.SelectPlugins(type=type, namePatterns=name_patterns, executor=executor, trigger=trigger))
    if not r.deleted:
      raise Exception(r.notDeletedReason)

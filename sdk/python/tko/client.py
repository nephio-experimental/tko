import tko.plugin, tko.tko_pb2_grpc, tko.tko_pb2, grpc


class Client:
  def __init__(self, host=None):
    self.host = host if host is not None else tko.plugin.get_grpc_host()

  def __enter__(self):
    self.channel = grpc.insecure_channel(self.host)
    self.stub = tko.tko_pb2_grpc.APIStub(self.channel)
    return self

  def __exit__(self, *args):
    self.channel.close()

  def list_sites(self):
    return self.stub.listSites(tko.tko_pb2.ListSites())

  # TODO

import json, sys, os


class Inventory:
  def __init__(self):
    self.hosts = {}

  def add(self, name, host):
    self.hosts[name] = host

  def dump(self, out=sys.stdout):
    inventory = {'_meta': {'hostvars': self.hosts}}
    inventory['all'] = {'hosts': list(self.hosts.keys())}
    json.dump(inventory, out, separators=(',', ':'))

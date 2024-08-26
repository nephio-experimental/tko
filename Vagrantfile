Vagrant.configure("2") do |config|
    config.vm.box = "fedora/40-cloud-base"

    config.vm.define "tko"
    config.vm.hostname = "tko"

    # Doesn't seem to work
    # config.vm.disk :disk, size: "8GB", primary: true

    config.vm.provider :libvirt do |libvirt|
        libvirt.cpus = 6
        libvirt.memory = 8192
        libvirt.machine_virtual_size = 8 # GB
    end

    config.vm.provider :virtualbox do |virtualbox|
        virtualbox.cpus = 6
        virtualbox.memory = 8192
        # TODO: how to increase disk size to 8GB?

        # https://www.mkwd.net/improve-vagrant-performance/
        virtualbox.customize ["modifyvm", :id, "--ioapic", "on"]
        virtualbox.customize ["modifyvm", :id, "--natdnshostresolver1", "on"]
        virtualbox.customize ["modifyvm", :id, "--natdnsproxy1", "on"]
    end

    # Increase partition size to fill our increased volume size
    config.vm.provision "shell", inline: <<-EOF
        growpart /dev/vda 4
        btrfs filesystem resize max /
    EOF

    # As an alternative to using "vagrant rsync-auto" you can mount as NFS
    #config.vm.synced_folder '.', '/vagrant', type: "nfs"

    config.vm.network :forwarded_port, guest: 50050, host: 60050 # gRPC API
    config.vm.network :forwarded_port, guest: 50051, host: 60051 # web GUI

    config.vm.network :forwarded_port, guest: 30050, host: 60060 # Kind gRPC API
    config.vm.network :forwarded_port, guest: 30051, host: 60061 # Kind web GUI

    config.vm.provision :shell, path: "scripts/install-vagrant", privileged: false
end

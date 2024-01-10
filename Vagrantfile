Vagrant.configure("2") do |config|
    config.vm.define "tko"
    config.vm.box = "fedora/39-cloud-base"

    # As an alternative to using "vagrant rsync-auto" you can mount as NFS
    #config.vm.synced_folder '.', '/vagrant', type: "nfs"

    # gRPC API port
    config.vm.network "forwarded_port", guest: 50050, host: 60050

    # Web GUI port
    config.vm.network "forwarded_port", guest: 50051, host: 60051

    config.vm.provision "shell", path: "scripts/install-vagrant", privileged: false
    # We need to reload for user permissions for Docker to take hold
    config.vm.provision :reload
    config.vm.provision "shell", inline: "/vagrant/scripts/test", privileged: false

    config.vm.provider :libvirt do |libvirt|
        libvirt.cpus = 4
        libvirt.memory = 8192
    end
end

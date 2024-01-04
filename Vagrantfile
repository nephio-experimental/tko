$root_script = <<-'SCRIPT'
/vagrant/scripts/install-system-dependencies-fedora -y
/vagrant/scripts/install-system-dependencies
/vagrant/scripts/setup-postgresql
SCRIPT

$user_script1 = <<-'SCRIPT'
echo "export PATH=$HOME/go/bin:$PATH" >> "$HOME/.bashrc"
SCRIPT

$user_script2 = <<-'SCRIPT'
/vagrant/scripts/install-python-dependencies
/vagrant/scripts/build
/vagrant/scripts/install-systemd-services
/vagrant/scripts/test
SCRIPT

Vagrant.configure("2") do |config|
    config.vm.define "tko"
    config.vm.box = "fedora/39-cloud-base"
    config.vm.provision "shell", inline: $root_script
    config.vm.provision "shell", inline: $user_script1, privileged: false
    config.vm.provision "shell", inline: $user_script2, privileged: false
    config.vm.network "forwarded_port", guest: 50051, host: 60051
    config.vm.provider :libvirt do |libvirt|
        libvirt.cpus = 4
        libvirt.memory = 4096
    end
end

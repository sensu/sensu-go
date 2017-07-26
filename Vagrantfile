# -*- mode: ruby -*-
# vi: set ft=ruby :
Vagrant.configure("2") do |config|
  config.vm.box = "ubuntu/xenial64"
  config.vm.synced_folder ".", "/sensu-go"

  config.vm.provision "shell", inline: <<-SHELL
    apt-get update
    apt-get install -y ruby ruby-dev build-essential rpm rpmlint
    gem install --no-ri --no-rdoc fpm
    gem install --no-ri --no-rdoc pleaserun
  SHELL
end

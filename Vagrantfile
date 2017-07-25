# -*- mode: ruby -*-
# vi: set ft=ruby :
Vagrant.configure("2") do |config|
  config.vm.box = "ubuntu/trusty64"
  config.vm.synced_folder ".", "/vagrant_data"

  config.vm.provision "shell", inline: <<-SHELL
    apt-get install -y ruby ruby-dev build-essential
    gem install --no-ri --no-rdoc fpm
    gem install --no-ri --no-rdoc pleaserun
  SHELL
end

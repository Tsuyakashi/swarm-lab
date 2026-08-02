if File.exist?(".env")
    File.foreach(".env") do |line|
        next if line.strip.empty? || line.start_with?("#")
        key, value = line.strip.split('=', 2)
        ENV[key] = value.gsub(/\A['"]|['"]\z/, '')
    end
end


ENV['VAGRANT_SERVER_URL'] = 'https://vagrant.elab.pro'

NODES = {
    "prod-node" => { hostname: "prod", ip: "192.168.56.10", memory: 1024, cpus: 1 },
    "stage-node" => { hostname: "stage", ip: "192.168.56.11", memory: 1024, cpus: 1 },
    "dev-node" => { hostname: "dev", ip: "192.168.56.12", memory: 1024, cpus: 1 }
}

PROD_IP = NODES["prod-node"][:ip]

Vagrant.configure("2") do |config|
    

    config.vm.box          = "bento/ubuntu-24.04"
    config.vm.boot_timeout = 300


    NODES.each do |name, cfg|
        config.vm.define name do |node|
            node.vm.hostname = cfg[:hostname]
            node.vm.network "private_network", ip: cfg[:ip]

            config.vm.synced_folder ".", "/vagrant", disabled: true    


            # Libvirt (Linux host)
            node.vm.provider "libvirt" do |lv|
                lv.memory            = cfg[:memory]
                lv.cpus              = cfg[:cpus]
                lv.storage_pool_name = "images"
            end
        
            # VirtualBox (Windows host)
            node.vm.provider "virtualbox" do |vb|
                vb.memory       = cfg[:memory]
                vb.cpus         = cfg[:cpus]
                vb.linked_clone = true
            end

            
            if name == NODES.keys.last
                node.vm.provision "ansible" do |ansible|
                    ansible.playbook          = "ansible/site.yml"
                    ansible.galaxy_role_file  = "ansible/requirements.yml"
                    ansible.limit             = "all"
                    ansible.groups = {
                        "swarm_managers" => ["prod-node"],
                        "swarm_workers"  => ["stage-node", "dev-node"],
                        "all:vars" => { "ansible_python_interpreter" => "/usr/bin/python3" }
                    }
                    ansible.host_vars = {
                        "prod-node"  => { "tag_name" => "prod" },
                        "stage-node" => { "tag_name" => "stage" },
                        "dev-node"   => { "tag_name" => "dev" }
                    }
                    ansible.extra_vars = {
                        base_registry: ENV['BASE_REGISTRY'],
                        github_repo:   ENV['GITHUB_REPO'],
                        github_pat:    ENV['GITHUB_PAT']
                    }
                end
            end
        end
    end
end

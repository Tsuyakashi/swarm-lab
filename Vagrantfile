if File.exist?(".env")
    File.foreach(".env") do |line|
        next if line.strip.empty? || line.start_with?("#")
        key, value = line.strip.split('=', 2)
        ENV[key] = value.gsub(/\A['"]|['"]\z/, '')
    end
end


ENV['VAGRANT_SERVER_URL'] = 'https://vagrant.elab.pro'

NODES = {
    "manager-node" => { hostname: "manager", ip: "192.168.56.10", memory: 1024, cpus: 1 },
    "worker1-node" => { hostname: "worker1", ip: "192.168.56.11", memory: 1024, cpus: 1 },
    "worker2-node" => { hostname: "worker2", ip: "192.168.56.12", memory: 1024, cpus: 1 }
}

MANAGER_IP = NODES["manager-node"][:ip]

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

            if name == "manager-node"
                node.vm.provision "file", source: "docker-compose.yml", destination: "/tmp/docker-compose.yml"

                node.vm.provision "configure_manager", type: "shell" do |s|
                    s.path = "scripts/manager.sh"
                    s.binary = true
                    s.env = {
                        "MANAGER_IP" => MANAGER_IP,
                        "BASE_REGISTRY" => ENV['BASE_REGISTRY']
                    }
                end
            else
                node.vm.provision "configure_worker", type: "shell" do |s|
                    s.path = "scripts/worker.sh"
                    s.binary = true
                    s.env = {}
                end
            end
        end
    end

    config.trigger.after :up do |trigger|
        trigger.ruby do |env, machine|
            token = nil
            30.times do
                out = `vagrant ssh manager-node -c "sudo docker info 2>/dev/null | grep -q 'Swarm: active' && sudo docker swarm join-token -q worker" 2>/dev/null`.strip
                unless out.empty?
                    token = out
                    break
                end
                sleep 5
            end
            raise "manager swarm not ready after timeout" if token.nil?

            NODES.each do |name, _cfg|
                next if name == "manager-node"
                state = `vagrant status #{name} --machine-readable`.lines.grep(/,state,/).last
                next unless state&.include?(",running")
                puts "== joining #{name} to swarm =="
                system("vagrant ssh #{name} -c \"...\"")
            end
        end
    end
end

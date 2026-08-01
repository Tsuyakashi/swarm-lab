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

            if name == "prod-node"
                node.vm.provision "file", source: "docker-compose.yml", destination: "/tmp/docker-compose.yml"
                node.vm.provision "file", source: "docker-compose.stage.yml", destination: "/tmp/docker-compose.stage.yml"
                node.vm.provision "file", source: "docker-compose.dev.yml", destination: "/tmp/docker-compose.dev.yml"

                node.vm.provision "configure_prod", type: "shell" do |s|
                    s.path = "scripts/manager.sh"
                    s.binary = true
                    s.env = {
                        "PROD_IP" => PROD_IP,
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
            next unless machine.name.to_s == NODES.keys.last  # выполняем один раз, после последней ноды
    
            token = nil
            30.times do
                out = `vagrant ssh prod-node -c "sudo docker info 2>/dev/null | grep -q 'Swarm: active' && sudo docker swarm join-token -q worker" 2>/dev/null`.strip
                unless out.empty?
                    token = out
                    break
                end
                sleep 5
            end
            raise "prod swarm not ready after timeout" if token.nil?
    
            NODES.each do |name, _cfg|
                next if name == "prod-node"
                state = `vagrant status #{name} --machine-readable`.lines.grep(/,state,/).last
                next unless state&.include?(",running")
                puts "== joining #{name} to swarm =="
                system("vagrant ssh #{name} -c \"sudo docker info 2>/dev/null | grep -q 'Swarm: active' || sudo docker swarm join --token #{token} #{PROD_IP}:2377\"")
            end
    
            puts "== deploying stack =="
            system("vagrant ssh prod-node -c \"sudo docker node update --label-add TAG=prod prod\"")
            system("vagrant ssh prod-node -c \"sudo docker node update --label-add TAG=stage stage\"")
            system("vagrant ssh prod-node -c \"sudo docker node update --label-add TAG=dev dev\"")
            system("vagrant ssh prod-node -c \"export BASE_REGISTRY=#{ENV['BASE_REGISTRY']}; sudo -E docker stack deploy --with-registry-auth -c /tmp/docker-compose.yml prod\"")
            system("vagrant ssh prod-node -c \"export BASE_REGISTRY=#{ENV['BASE_REGISTRY']}; sudo -E docker stack deploy --with-registry-auth -c /tmp/docker-compose.stage.yml stage\"")
            system("vagrant ssh prod-node -c \"export BASE_REGISTRY=#{ENV['BASE_REGISTRY']}; sudo -E docker stack deploy --with-registry-auth -c /tmp/docker-compose.dev.yml dev\"")
        end
    end
end

# swarm-lab

Minimal two-node Docker Swarm lab: provisioning, cluster bootstrap, and a rolling-update
stack deployed automatically via `vagrant up`.

## Architecture

```
┌──────────────────┐         overlay network         ┌─────────────────┐
│  manager-node    │◄───────────────────────────────►│  worker-node    │
│  192.168.56.11   │                                 │  192.168.56.12  │
│                  │                                 │                 │
│  nginx (:80)     │                                 │  nodejs-app     │
│  python-app      │                                 │  go-app         │
└──────────────────┘                                 └─────────────────┘
```

- **manager-node** — swarm leader, runs `nginx` (reverse proxy, entrypoint on :80) and `python-app`
- **worker-node** — runs `nodejs-app` and `go-app`
- Services communicate over a Swarm overlay network (`web-network`); nginx proxies
  `/go/`, `/nodejs/`, `/python/` to the respective backend

## Stack

| Component | Purpose |
|---|---|
| Vagrant + libvirt/VirtualBox | VM provisioning (auto-selects provider per host OS) |
| Docker Swarm | Orchestration, rolling updates |
| GitHub Actions | Per-service builds on change, push to GHCR |
| nginx | Reverse proxy / entrypoint |
| Go / Node.js / Python | Toy backend services with `/health` endpoints |

## Quickstart

```bash
cp .env.exmaple .env   # set BASE_REGISTRY
vagrant up
```

This will:
1. Boot `manager-node` and `worker-node`
2. Install Docker on both
3. Init the swarm on the manager, join the worker automatically (via a Vagrant trigger
   that polls for the join token)
4. Pull images from `BASE_REGISTRY` and deploy the stack (`docker stack deploy`)

Check cluster state:

```bash
vagrant ssh manager-node -c "docker node ls"
vagrant ssh manager-node -c "docker ps"
```

Check services:

```bash
curl 192.168.56.11/go/health
curl 192.168.56.11/nodejs/health
curl 192.168.56.11/python/health
```

## CI/CD

`.github/workflows/ci.yml` detects which service directory changed (`dorny/paths-filter`)
and builds/pushes only that image to GHCR, tagged `<service>-<sha>` and `<service>-latest`.

Deploy-on-push to Swarm is not wired up yet — currently `vagrant up` always pulls
`*-latest` on provision.

## Roadmap

- [ ] CD: trigger `docker service update` on the manager after a successful CI build
- [ ] Basic monitoring (cAdvisor/Prometheus or similar)
- [ ] TLS on nginx entrypoint
- [ ] Split into its own repo / link from devops-handbook

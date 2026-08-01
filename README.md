# swarm-lab

Minimal three-node Docker Swarm lab: provisioning, cluster bootstrap, and per-environment
stacks (prod/stage/dev) deployed automatically via `vagrant up`.

## Architecture

```
┌──────────────────┐   ┌──────────────────┐   ┌──────────────────┐
│   prod-node      │   │   stage-node     │   │   dev-node       │
│   192.168.56.10  │   │   192.168.56.11  │   │   192.168.56.12  │
│   (swarm leader) │   │                  │   │                  │
│                  │   │                  │   │                  │
│  nginx (:80)     │   │  nginx (:8080)   │   │  nginx (:8081)   │
│  python-app      │   │  python-app      │   │  python-app      │
│  nodejs-app      │   │  nodejs-app      │   │  nodejs-app      │
│  go-app          │   │  go-app          │   │  go-app          │
└──────────────────┘   └──────────────────┘   └──────────────────┘
```

- All three nodes join a single Swarm cluster; `prod-node` is the manager/leader.
- Each node is labeled `TAG=prod|stage|dev`, and each stack (`prod`, `stage`, `dev`)
  is placement-constrained to its matching node via `node.labels.TAG==<env>`.
- Each stack is fully self-contained: its own nginx + all three backends, its own
  overlay network (`<stack>_web-network`).
- nginx in each stack proxies `/go/`, `/nodejs/`, `/python/` to the respective backend.
- Thanks to Swarm's routing mesh, any published port is reachable from **any** node's
  IP — e.g. `192.168.56.10:8080` and `192.168.56.11:8080` both hit the `stage` stack.
  Below, all examples use `192.168.56.10` for convenience.

## Stack

| Component                     | Purpose                                                   |
| ----------------------------- | --------------------------------------------------------- |
| Vagrant + libvirt/VirtualBox  | VM provisioning (auto-selects provider per host OS)       |
| Docker Swarm                  | Orchestration, rolling updates, placement by node label   |
| GitHub Actions                | Per-service, per-environment builds, push to GHCR         |
| nginx                         | Reverse proxy / entrypoint                                |
| Go / Node.js / Python         | Toy backend services with `/health` endpoints             |

## Quickstart

```bash
cp .env.example .env   # set BASE_REGISTRY
vagrant up
```

This will:

1. Boot `prod-node`, `stage-node`, `dev-node`
2. Install Docker on all three (+ compose plugin on `prod-node`)
3. Init the swarm on `prod-node`, join the other two as workers automatically
   (via a Vagrant trigger that polls for the join token)
4. Label each node (`TAG=prod|stage|dev`) and deploy all three stacks
   (`docker stack deploy`) from `prod-node`, pulling images from `BASE_REGISTRY`

Check cluster state:

```bash
vagrant ssh prod-node -c "docker node ls"
vagrant ssh prod-node -c "docker service ls"
```

Check services (routing mesh — any node IP works):

```bash
curl 192.168.56.10/go/health       # prod
curl 192.168.56.10:8080/go/health  # stage
curl 192.168.56.10:8081/go/health  # dev
```

## CI/CD

Three workflows build and push per-service images to GHCR, tagged `<service>-<sha>`
and `<service>-<env>`:

| Workflow         | Trigger                            | Tag               |
| ---------------- | ---------------------------------- | ----------------- |
| `ci.yml`         | push to `main`                     | `-latest` (prod)  |
| `ci.stage.yml`   | pull request                       | `-stage`          |
| `ci.dev.yml`     | push to any branch except `main`   | `-dev`            |

All three use `dorny/paths-filter` to build only the service directories that changed.

Deploy is handled entirely by the `Vagrantfile` trigger on `vagrant up` — it pulls
`*-latest`/`*-stage`/`*-dev` per stack and runs `docker stack deploy`.

## Roadmap

- [ ] Basic monitoring (cAdvisor/Prometheus or similar)
- [ ] TLS on nginx entrypoint
- [ ] Split into its own repo / link from devops-handbook

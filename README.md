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
- nginx in each stack proxies `/go/`, `/nodejs/`, `/python/` to the respective backend,
  and serves a landing page listing all services and their `/health` endpoints.
- Thanks to Swarm's routing mesh, any published port is reachable from **any** node's
  IP — e.g. `192.168.56.10:8080` and `192.168.56.11:8080` both hit the `stage` stack.
  Below, all examples use `192.168.56.10` for convenience.

## Stack

| Component                     | Purpose                                                     |
| ------------------------------ | ------------------------------------------------------------ |
| Vagrant + libvirt/VirtualBox   | VM provisioning (auto-selects provider per host OS)          |
| Ansible                        | Docker install, Swarm bootstrap, node labeling, stack deploy, self-hosted runner setup |
| Docker Swarm                   | Orchestration, rolling updates with rollback, placement by node label |
| GitHub Actions (self-hosted runner) | Build, promote, deploy, and verify per changed service |
| nginx                          | Reverse proxy / entrypoint + landing page                    |
| Go / Node.js / Python          | Toy backend services, each with HTML rendering and a `/health` endpoint |

### Applications

- **Go** — stdlib `net/http`, HTML rendered via embedded `html/template` (`go:embed`),
  structured logging with `log/slog`, health check doubles as the container's own
  `HEALTHCHECK` binary flag (`-healthcheck`).
- **Node.js** — Express + EJS views, `morgan` request logging (skips `/health`),
  includes a `/quote` route serving a random quote.
- **Python** — Flask behind Gunicorn (4 workers), renders `index.html` with a
  timestamp and current 1-minute system load average.
- **nginx** — reverse proxy for the three backends, plus a static landing page and
  a lightweight `/health` endpoint of its own.

## Quickstart

```bash
cp .env.example .env   # set BASE_REGISTRY, GITHUB_REPO, GITHUB_PAT
vagrant up
```

This will:

1. Boot `prod-node`, `stage-node`, `dev-node`.
2. Install Docker on all three nodes (Ansible `docker` role).
3. Init the Swarm on `prod-node`, join the other two as workers.
4. Label each node (`TAG=prod|stage|dev`).
5. Deploy all three stacks (`docker stack deploy`) from `prod-node`, pulling images
   from `BASE_REGISTRY`.
6. Register a self-hosted GitHub Actions runner on `prod-node` (Ansible
   `github-runner` role), which the CI/CD pipeline uses for deployment.

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

A single workflow, `.github/workflows/pipeline.yml`, handles build, promotion,
deployment, and verification for all three environments:

| Trigger                                      | Resolved stack | Image tag |
| -------------------------------------------- | -------------- | --------- |
| Push to any branch except `main`             | `dev`          | `dev`     |
| Pull request opened / synchronized           | `stage`        | `stage`   |
| Pull request merged into `main`              | `prod`         | `latest`  |

Pipeline stages:

1. **detect-changes** — uses `dorny/paths-filter` to find which service directories
   (`python/`, `go/`, `nodejs/`, `nginx/`) changed, and resolves the target stack,
   image tag, and commit SHA from the event type.
2. **build** — for `dev` pushes, builds and pushes each changed service to GHCR,
   tagged both `<service>-<sha>` and `<service>-dev`.
3. **promote** — for `stage`/`prod`, retags the existing `<service>-<sha>` image as
   `<service>-stage` / `<service>-latest` via `docker buildx imagetools create`,
   with no rebuild.
4. **deploy** — runs on the self-hosted runner; pulls the resolved image and runs
   `docker service update --with-registry-auth --force` against the matching stack
   service.
5. **verify** — waits for the updated service to converge (desired vs. running
   replica count), then hits its `/health` endpoint (via nginx for the `nginx`
   service, via `/<service>/health` for the others) and fails the run on a non-200
   response.

A `concurrency` group (per workflow + branch/ref) prevents overlapping pipeline runs
from racing each other on the same stack.

## Roadmap

- [ ] Basic monitoring (cAdvisor/Prometheus or similar)
- [ ] TLS on nginx entrypoint
- [ ] Split into its own repo / link from devops-handbook

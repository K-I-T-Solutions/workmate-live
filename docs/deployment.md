# Deployment auf workmate-01

Portal auf dem Server, Agent auf dem Streaming-Rechner. Der Agent baut die
Verbindung auf (`obs.mode: agent`), der Streaming-Rechner braucht also keinen
eingehenden Port.

Gebaut wird in GitHub Actions, nicht auf dem Server: Der Workflow schiebt
Backend- und Frontend-Image in die GitHub Container Registry, der Server zieht
sie nur noch.

```
  git push main
        │
        ▼
  GitHub Actions ──build──► ghcr.io/k-i-t-solutions/workmate-live-{backend,frontend}
        │                                    │
        └──────ssh: compose pull & up────────┘
                        │
                        ▼
                  workmate-01  ──Caddy──►  live.workmate.kit-it-koblenz.de
                        ▲
                        │ wss:// (der Agent wählt sich ein)
                  Streaming-PC + OBS
```

## Zielumgebung

| | |
|---|---|
| Server | `workmate-01` (77.42.17.200), Ubuntu 24.04, x86_64 |
| Reverse Proxy | Caddy 2 (`workmate_caddy`), Port 80/443, im `core_network` |
| Caddyfile | `/srv/workmate/reverse-proxy/Caddyfile` |
| Projektpfad | `/srv/workmate/projects/workmate-live` |
| Domain | `live.workmate.kit-it-koblenz.de` |
| Registry | `ghcr.io/k-i-t-solutions/workmate-live-{backend,frontend}` |

Das Schema entspricht `workmate-event` und `workmate-access`: Git-Repo im
Projektordner, Container ohne Port-Mapping im `core_network`, Caddy davor.

### Service-Namen im geteilten Netzwerk

`core_network` ist extern und wird von mehreren Compose-Projekten benutzt.
Docker Compose vergibt jedem Container **zwei** DNS-Aliase: den Container-Namen
und den *Service*-Namen aus der Compose-Datei. Zwei Projekte mit gleichem
Service-Namen erzeugen damit denselben Alias im selben Netzwerk — Docker
verteilt Anfragen dann per Round-Robin auf beide Container.

`workmate_access_backend` belegt dort bereits den Alias `backend`. Die Services
heißen deshalb `workmate-live-backend` und `workmate-live-frontend`, und
`portal/frontend/nginx.conf` proxyt entsprechend auf `workmate-live-backend:8080`.

Prüfen, bevor ein neuer Service dazukommt:

```bash
ssh workmate-01 'for c in $(docker network inspect core_network \
  --format "{{range .Containers}}{{.Name}} {{end}}"); do \
  echo -n "$c: "; docker inspect $c \
  --format "{{index .NetworkSettings.Networks \"core_network\" \"Aliases\"}}"; done'
```

## 1. DNS in Cloudflare

Die Zone `kit-it-koblenz.de` liegt bei Cloudflare (`athena`/`harvey.ns.cloudflare.com`).
Alle bestehenden Subdomains zeigen **direkt** auf die Server-IP, laufen also
ohne Cloudflare-Proxy (graue Wolke).

Der Eintrag:

| Feld | Wert |
|---|---|
| Type | `A` |
| Name | `live.workmate` |
| IPv4 | `77.42.17.200` |
| Proxy status | **DNS only** (graue Wolke) |
| TTL | Auto |

Per API anlegen (Token in `~/.credentials/.cloudflare`, kann Zonen und DNS
lesen und schreiben):

```bash
set -a; . ~/.credentials/.cloudflare; set +a

ZID=$(curl -s "https://api.cloudflare.com/client/v4/zones?name=kit-it-koblenz.de" \
  -H "Authorization: Bearer $CLOUDFLARE_API_TOKEN" | jq -r '.result[0].id')

curl -s -X POST "https://api.cloudflare.com/client/v4/zones/$ZID/dns_records" \
  -H "Authorization: Bearer $CLOUDFLARE_API_TOKEN" \
  -H "Content-Type: application/json" \
  --data '{"type":"A","name":"live.workmate.kit-it-koblenz.de",
           "content":"77.42.17.200","proxied":false,"ttl":1}' | jq '.success'
```

Prüfen:

```bash
dig +short A live.workmate.kit-it-koblenz.de   # erwartet: 77.42.17.200
```

### Reihenfolge beachten

**Erst den DNS-Eintrag anlegen, dann Caddy neu laden.** Lädt Caddy die neue
Domain, bevor der Eintrag existiert, scheitert die ACME-Prüfung mit `NXDOMAIN`
— und Let's Encrypt cacht diese negative Antwort. Die Zone hat eine
SOA-Minimum-TTL von 1800 s, das Zertifikat kommt dann also bis zu 30 Minuten
später als nötig. Caddy versucht es von selbst weiter, es geht nichts kaputt;
es dauert nur.

Fortschritt verfolgen:

```bash
ssh workmate-01 'docker logs --tail 200 workmate_caddy 2>&1 \
  | grep -i "live.workmate" | grep -vi stacktrace | tail -5'
```

### Warum DNS only statt Proxy

- Caddy holt sein Zertifikat per ACME HTTP-01. Ohne Proxy funktioniert das
  sofort und ohne Sonderfälle.
- Der Agent-Link ist eine **dauerhafte** WebSocket-Verbindung. Cloudflare
  kappt im Free-Plan inaktive WebSockets; der Ping-Intervall des Agents (25 s)
  hält sie zwar offen, aber die Abhängigkeit ist unnötig.
- Es bleibt konsistent mit den vier bestehenden Subdomains.

Mit Proxy (orange) wäre zusätzlich nötig: SSL/TLS-Modus **Full (strict)**,
sonst scheitert entweder ACME oder es entsteht eine Redirect-Schleife.

## 2. Registry und Deploy-Zugang einrichten (einmalig)

### 2.1 Erster Build

Der Workflow läuft bei jedem Push auf `main`:

```bash
gh workflow run "Build & Push Images"     # oder einfach pushen
gh run watch
```

### 2.2 Images sichtbar machen

Neue GHCR-Packages sind **privat**, auch wenn das Repo öffentlich ist. Da das
Repo ohnehin public ist, ist der einfachste Weg, die Packages ebenfalls auf
public zu stellen — dann braucht der Server keinen Registry-Login:

GitHub → Organisation `K-I-T-Solutions` → Packages → `workmate-live-backend`
→ Package settings → Change visibility → **Public**. Dasselbe für
`workmate-live-frontend`.

Alternativ privat lassen und den Server anmelden:

```bash
# PAT (classic) mit Scope read:packages
ssh workmate-01 'echo <PAT> | docker login ghcr.io -u commanderphu --password-stdin'
```

### 2.3 Deploy-Key

Schlüsselpaar erzeugen (ohne Passphrase, damit Actions ihn nutzen kann):

```bash
ssh-keygen -t ed25519 -f ~/.ssh/workmate_live_deploy -N "" -C "gh-actions-deploy"
ssh-copy-id -i ~/.ssh/workmate_live_deploy.pub workmate-01
```

Als Repository-Secrets hinterlegen:

```bash
gh secret set DEPLOY_SSH_KEY -R K-I-T-Solutions/workmate-live < ~/.ssh/workmate_live_deploy
gh secret set DEPLOY_HOST    -R K-I-T-Solutions/workmate-live --body "77.42.17.200"
gh secret set DEPLOY_USER    -R K-I-T-Solutions/workmate-live --body "joshua"
gh secret set DEPLOY_PATH    -R K-I-T-Solutions/workmate-live --body "/srv/workmate/projects/workmate-live"
```

Fehlen `DEPLOY_SSH_KEY` oder `DEPLOY_HOST`, überspringt der Workflow den
Deploy-Schritt und baut nur die Images.

## 3. Server vorbereiten (einmalig)

### 3.1 Platz schaffen

Der Server läuft auf 81 % (14 GB frei), der alte Docker-Build-Cache belegt
davon knapp 50 GB. Da künftig nicht mehr auf dem Server gebaut wird, kann er
weg:

```bash
ssh workmate-01 'docker builder prune -af'     # gibt ~49 GB frei
```

### 3.2 Repo klonen

```bash
ssh workmate-01
cd /srv/workmate/projects
git clone https://github.com/K-I-T-Solutions/workmate-live.git
cd workmate-live
```

Auf dem Server gebraucht werden nur `docker-compose.prod.yml` und die Configs —
der Quellcode liegt lediglich mit im Ordner, weil der Deploy-Schritt `git pull`
macht.

### 3.3 Portal konfigurieren

`portal/backend/config/portal.docker.yaml` wird nach `/app/config/portal.yaml`
gemountet. Sie ist gitignored und muss auf dem Server angelegt werden:

```yaml
server:
    address: 0.0.0.0
    port: 8080
    timeouts:
        read: 10s
        write: 10s
        shutdown: 5s
auth:
    # openssl rand -hex 32
    jwt_secret: "<JWT_SECRET>"
    token_duration: 24h0m0s
    default_user:
        username: admin
        password: "<ADMIN_PASSWORT>"
agent:
    # Leer: der Agent sitzt hinter NAT und wird nicht gepollt,
    # sein Status kommt über den Link.
    url: ""
    polling_interval: 3s
    timeout: 5s
    # openssl rand -hex 32 — identisch zum api_key im Agent
    api_key: "<AGENT_API_KEY>"
    primary_id: ""
    command_timeout: 10s
obs:
    mode: agent
twitch:
    enabled: false
youtube:
    enabled: false
automation:
    rules_file: config/rules.yaml
storage:
    type: sqlite
    path: /app/data/portal.db
logging:
    level: info
    format: text
```

```bash
chmod 600 portal/backend/config/portal.docker.yaml
```

### 3.4 Caddy-Block ergänzen

An `/srv/workmate/reverse-proxy/Caddyfile` anhängen:

```caddy
live.workmate.kit-it-koblenz.de {
    reverse_proxy workmate_live_ui:80
}
```

Mehr ist nicht nötig: Caddy erkennt WebSocket-Upgrades selbst und reicht sie
durch — sowohl `/ws` (Browser) als auch `/ws/agent` (Agent). Das nginx im
Frontend-Container verteilt intern weiter an das Backend.

Neu laden, ohne Caddy neu zu starten:

```bash
ssh workmate-01 'docker exec workmate_caddy caddy reload --config /etc/caddy/Caddyfile'
```

## 4. Erster Start

```bash
ssh workmate-01 'cd /srv/workmate/projects/workmate-live && \
  docker compose -f docker-compose.prod.yml pull && \
  docker compose -f docker-compose.prod.yml up -d'

ssh workmate-01 'docker logs -f workmate_live_backend'
```

Erwartete Zeilen:

```
Agent link enabled at /ws/agent
No agent URL configured, relying on agent link for status
OBS control routed through agent link
```

## 5. Laufender Betrieb

Ab jetzt genügt ein Push auf `main`: Actions baut beide Images, schiebt sie
nach ghcr.io und rollt per SSH aus (`git pull` → `compose pull` → `up -d`).
Danach prüft der Workflow `/health` und bricht mit den letzten 50 Logzeilen ab,
falls das Portal nicht hochkommt.

Manuell auslösen:

```bash
gh workflow run "Build & Push Images" -f deploy=true
```

Auf ein bestimmtes Image zurückrollen:

```bash
ssh workmate-01 'cd /srv/workmate/projects/workmate-live && \
  BACKEND_IMAGE=ghcr.io/k-i-t-solutions/workmate-live-backend:sha-<commit> \
  docker compose -f docker-compose.prod.yml up -d workmate-live-backend'
```

Jeder Build wird zusätzlich als `sha-<vollständiger-commit>` getaggt, ein
Git-Tag `v1.2.3` erzeugt außerdem `1.2.3` und `1.2`.

## 6. Agent auf dem Streaming-Rechner

`~/.config/workmate-agent/config.yaml` (oder `/etc/workmate-agent/config.yaml`):

```yaml
server:
  address: "127.0.0.1"
  port: 8787
  timeouts:
    read: 5s
    write: 5s
    shutdown: 5s

health:
  polling_interval: 2s
  checks:
    gpu: true
    audio: true
    video: true
    obs: true

obs:
  enabled: true
  host: "127.0.0.1"           # OBS bleibt auf localhost
  port: 4455
  password: "<OBS_PASSWORT>"  # OBS: Werkzeuge → WebSocket-Servereinstellungen
  reconnect_delay: 5s

portal:
  enabled: true
  url: "https://live.workmate.kit-it-koblenz.de"
  api_key: "<AGENT_API_KEY>"   # identisch zum Portal
  agent_id: ""                 # leer = Hostname
  timeout: 10s
  retry_attempts: 3
  retry_delay: 5s
```

`https://` verwenden — der Agent leitet daraus `wss://.../ws/agent` ab.

Als systemd-Service (siehe `agent/workmate-live-agent.service`):

```bash
sudo systemctl enable --now workmate-live-agent
journalctl -u workmate-live-agent -f
```

Erwartete Zeilen:

```
obs control enabled for 127.0.0.1:4455
link: connecting to portal at https://live.workmate.kit-it-koblenz.de
obs: connected to 127.0.0.1:4455
link: connected to portal
```

## 7. Prüfen

Serverseitig im Log:

```
agentlink: agent <hostname> connected
agentlink: agent <hostname> ready (host=<hostname> version=... obs=true)
```

Über die API:

```bash
TOKEN=$(curl -s -X POST https://live.workmate.kit-it-koblenz.de/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"<ADMIN_PASSWORT>"}' | jq -r .token)

curl -s -H "Authorization: Bearer $TOKEN" \
  https://live.workmate.kit-it-koblenz.de/api/agents | jq
```

Erwartet: ein Eintrag mit `"obs_local": true`, `"obs_active": true`,
`"connected": true`.

In der UI stehen die verbundenen Agents unter **Einstellungen → Agent**.

## Sicherheitshinweise

- `/ws/agent` ist nach dem Ausrollen öffentlich erreichbar. Der `api_key` ist
  die einzige Hürde davor — einen echten Zufallswert verwenden
  (`openssl rand -hex 32`), niemals ein selbst ausgedachtes Wort. Wer den
  Schlüssel hat, kann OBS steuern.
- `jwt_secret` und `api_key` gehören auf dasselbe Schutzniveau. Beide stehen im
  Klartext in `portal.docker.yaml` — gitignored und `chmod 600`.
- Das Standardpasswort `changeme` vor dem ersten öffentlichen Start ersetzen.
- Der Deploy-Key hat Shell- und Docker-Zugriff auf workmate-01. Er gehört nur
  in die Repository-Secrets, nicht in einen Branch oder eine Datei.

## Cloudflare-Hinweis für später

Sollte die Subdomain doch hinter den Cloudflare-Proxy wandern:

- SSL/TLS-Verschlüsselungsmodus auf **Full (strict)** stellen.
- WebSockets sind bei Cloudflare standardmäßig aktiv (Network → WebSockets).
- Kein „Caching Level: Cache Everything" auf dieser Subdomain — das würde die
  API-Antworten zwischenspeichern.

# Workmate Live

[![GitHub release](https://img.shields.io/github/v/release/K-I-T-Solutions/workmate-live?style=flat-square)](https://github.com/K-I-T-Solutions/workmate-live/releases)
[![License](https://img.shields.io/badge/license-Proprietary-red?style=flat-square)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.25.5-00ADD8?style=flat-square&logo=go)](https://golang.org/)
[![React](https://img.shields.io/badge/React-19.2.0-61DAFB?style=flat-square&logo=react)](https://reactjs.org/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.9.3-3178C6?style=flat-square&logo=typescript)](https://www.typescriptlang.org/)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat-square)](https://github.com/K-I-T-Solutions/workmate-live/pulls)

Eine plattformunabhängige, self-hosted Live-Streaming und Event-Broadcasting-Lösung, die Systemüberwachung, OBS Studio-Integration und Multi-Plattform-Streaming-Support für professionelle Events, Content Creator und hybride Veranstaltungen bietet.

## Überblick

Workmate Live besteht aus zwei Hauptkomponenten:

- **Agent**: Ein leichtgewichtiger Systemüberwachungsdienst, der Hardware- und Softwarestatus erfasst
- **Portal**: Eine Web-basierte Steuerungszentrale mit Backend-API und modernem Frontend

## Projektstruktur

```
workmate_live/
├── agent/              # Systemüberwachungs-Agent (Go)
├── portal/             # Web-Portal
│   ├── backend/       # Portal API-Server (Go)
│   └── frontend/      # Web-UI (React + TypeScript)
└── docs/              # Dokumentation
```

## Features

### Agent
- **Plattformunabhängig**: Linux, Windows und macOS (amd64 und arm64)
- **System-Monitoring**
  - GPU-Erkennung via `/dev/dri`
  - Audio-System-Status (PipeWire)
  - Video-Geräte-Scan (`/dev/video*`)
  - OBS Studio-Prozesserkennung
- **OBS-Fernsteuerung**: Steuert die lokale OBS-Instanz im Auftrag des Portals
- **Portal-Link**: Baut die Verbindung zum Portal selbst auf — funktioniert
  hinter NAT und Firewall, ohne eingehenden Port auf dem Streaming-Rechner
- **REST API** auf `127.0.0.1:8787`
- Konfigurierbare Polling-Intervalle

### Portal Backend
- **Agent-Integration**: Echtzeit-Status vom Systemagent, per HTTP-Polling oder
  über den Agent-Link
- **OBS Studio WebSocket**: Vollständige OBS-Steuerung — direkt oder über einen
  Agent auf einem entfernten Rechner
- **Streaming-Plattformen**:
  - Twitch-Integration
  - YouTube-Integration
- **Authentifizierung**: JWT-basierte Auth
- **Speicher**: SQLite-Datenbank
- **API-Server** auf `0.0.0.0:8080`

### Portal Frontend
- **Moderne Tech-Stack**:
  - React 19 mit TypeScript
  - Vite als Build-Tool
  - Tailwind CSS für Styling
  - Zustand für State Management
- **UI-Komponenten**: Radix UI Primitives
- **Icons**: Lucide React
- **Responsive Design**

## Technologien

### Backend
- **Sprache**: Go 1.25.5
- **Frameworks**:
  - Chi Router (`go-chi/chi`)
  - CORS Support (`go-chi/cors`)
  - OBS WebSocket (`andreykaipov/goobs`)
  - WebSocket (`gorilla/websocket`)
- **Konfiguration**: YAML

### Frontend
- **Framework**: React 19.2.0
- **Build**: Vite 7.2.4
- **Sprache**: TypeScript 5.9.3
- **Styling**: Tailwind CSS 4.1.18
- **State**: Zustand 5.0.9
- **UI Library**: Radix UI

## Installation

### Voraussetzungen
- Go 1.25.5 oder höher
- Node.js und npm (für Frontend)
- OBS Studio (optional, für Streaming-Features)

### Agent Setup

```bash
cd agent

# Konfiguration erstellen
cp config.example.yaml config.yaml

# Binary builden
go build -o workmate-live-agent cmd/workmate-live-agent/main.go

# Agent starten
./workmate-live-agent
```

### Portal Backend Setup

```bash
cd portal/backend

# Konfiguration erstellen
cp config/portal.example.yaml config/portal.yaml
# Bearbeiten Sie portal.yaml und ändern Sie:
# - JWT Secret (verwenden Sie: openssl rand -hex 32)
# - Default Admin-Passwort
# - OBS WebSocket-Einstellungen
# - API-Keys für Twitch/YouTube

# Binary builden
go build -o workmate-live-portal cmd/workmate-live-portal/main.go

# Portal starten
./workmate-live-portal
```

### Portal Frontend Setup

```bash
cd portal/frontend

# Dependencies installieren
npm install

# Development Server
npm run dev

# Production Build
npm run build
```

## Konfiguration

### Agent (`agent/config.yaml`)

```yaml
server:
  address: "127.0.0.1"
  port: 8787

health:
  polling_interval: 2s
  checks:
    gpu: true
    audio: true
    video: true
    obs: true
```

### Portal Backend (`portal/backend/config/portal.yaml`)

```yaml
server:
  address: "0.0.0.0"
  port: 8080

auth:
  jwt_secret: "CHANGE_THIS"
  token_duration: 24h
  default_user:
    username: "admin"
    password: "changeme"

agent:
  url: "http://127.0.0.1:8787"
  polling_interval: 3s

obs:
  mode: direct
  host: "192.168.178.100"
  port: 4455
  password: "your-obs-password"
  auto_reconnect: true
```

## OBS auf einem entfernten Rechner

Es gibt zwei Wege, OBS zu erreichen. Der Modus steht in `obs.mode` und lässt
sich auch in der Portal-UI unter **Einstellungen → OBS** umschalten.

### `direct` (Standard)

Das Portal verbindet sich selbst zum OBS-WebSocket. Einfach einzurichten, aber
OBS muss vom Portal aus erreichbar sein und der OBS-WebSocket kennt kein TLS —
also nur im LAN oder über ein VPN (WireGuard, Tailscale) verwenden.

### `agent`

Der Agent läuft auf dem Streaming-Rechner, verbindet sich dort lokal mit OBS
und baut **von sich aus** eine WebSocket-Verbindung zum Portal auf. Das Portal
schickt Kommandos durch diese Verbindung zurück.

Vorteile gegenüber `direct`:

- Der Streaming-Rechner braucht **keinen eingehenden Port** — funktioniert
  hinter NAT und Firewall.
- Das OBS-Passwort bleibt auf dem Streaming-Rechner.
- Die Verbindung nutzt das TLS des Portals (`https://` → `wss://`).
- Mehrere Rechner können sich gleichzeitig verbinden.

**Portal** (`portal/backend/config/portal.yaml`):

```yaml
agent:
  # Leer lassen: ein Agent hinter NAT kann nicht gepollt werden,
  # sein Status kommt über den Link.
  url: ""
  # Gemeinsames Geheimnis, z. B. via: openssl rand -hex 32
  api_key: "SHARED_SECRET"
  # Welcher Agent OBS steuert. Leer = erster verbundener Agent mit OBS.
  primary_id: ""
  command_timeout: 10s

obs:
  mode: agent
```

**Agent** (`agent/config.yaml`) auf dem Streaming-Rechner:

```yaml
obs:
  enabled: true
  host: "127.0.0.1"    # OBS bleibt auf localhost
  port: 4455
  password: "your-obs-password"
  reconnect_delay: 5s

portal:
  enabled: true
  url: "https://portal.example.com"
  api_key: "SHARED_SECRET"   # identisch zu agent.api_key im Portal
  agent_id: ""               # leer = Hostname
  retry_attempts: 3
  retry_delay: 5s
```

Der Agent hängt `/ws/agent` selbst an die URL an und wandelt `http://` bzw.
`https://` nach `ws://` bzw. `wss://` um.

Verbundene Agents stehen unter `GET /api/agents` und in der UI unter
**Einstellungen → Agent**. Beide Seiten verbinden nach einem Abriss selbständig
neu (der Agent mit wachsendem Backoff, gedeckelt auf 2 Minuten).

> **Reverse Proxy**: `/ws/agent` ist eine WebSocket-Route. Ein davorliegender
> Proxy muss die Header `Upgrade` und `Connection` durchreichen und braucht
> ein großzügiges Read-Timeout, sonst wird die Verbindung im Leerlauf gekappt.

## Deployment

Produktivbetrieb läuft über GitHub Actions: Der Workflow baut Backend- und
Frontend-Image, schiebt sie in die GitHub Container Registry und rollt sie per
SSH auf dem Server aus. Auf dem Server wird nichts gebaut.

```bash
docker compose -f docker-compose.prod.yml pull
docker compose -f docker-compose.prod.yml up -d
```

Vollständige Anleitung inklusive DNS, Caddy und Agent-Einrichtung:
**[docs/deployment.md](docs/deployment.md)**

Agent auf Windows, macOS oder Linux einrichten:
**[docs/agent-install.md](docs/agent-install.md)**

## API-Endpunkte

### Agent API (`http://127.0.0.1:8787`)
- `GET /health` - System-Health-Status
- `GET /info` - Agent-Informationen
- `GET /capabilities` - Erkannte Hardware-Fähigkeiten

### Portal API (`http://0.0.0.0:8080`)
- `POST /auth/login` - Authentifizierung
- `GET /agent/status` - Agent-Status abrufen
- `GET /agents` - Über den Link verbundene Agents
- `WS /ws` - WebSocket für Echtzeit-Updates (Browser, JWT)
- `WS /ws/agent` - Agent-Link (Agents, API-Key)
- `POST /obs/*` - OBS Studio-Steuerung
- `GET /twitch/*` - Twitch-Integration
- `GET /youtube/*` - YouTube-Integration

## Entwicklung

### Agent Entwicklung
```bash
cd agent
go run cmd/workmate-live-agent/main.go
```

### Portal Backend Entwicklung
```bash
cd portal/backend
go run cmd/workmate-live-portal/main.go
```

### Portal Frontend Entwicklung
```bash
cd portal/frontend
npm run dev
```

Der Dev-Server läuft auf `http://localhost:5173` (oder einem anderen Port, falls 5173 belegt ist).

## Sicherheit

- Ändern Sie das JWT-Secret in Produktion
- Verwenden Sie starke Passwörter
- Aktivieren Sie HTTPS für Produktionsumgebungen
- Speichern Sie API-Keys sicher (nicht in Git committen)
- Die `config.yaml` Dateien sind bereits in `.gitignore`

## Lizenz

Proprietäre Software - Alle Rechte vorbehalten

## Support

Bei Fragen oder Problemen erstellen Sie bitte ein Issue im Repository oder kontaktieren Sie das Entwicklungsteam.

## Über K.I.T Solutions

Entwickelt und gewartet von [K.I.T Solutions](https://github.com/K-I-T-Solutions)

## Contributing

Contributions sind willkommen! Bitte erstellen Sie einen Pull Request oder öffnen Sie ein Issue für Vorschläge und Verbesserungen.

1. Fork das Repository
2. Erstellen Sie einen Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Committen Sie Ihre Änderungen (`git commit -m 'Add some AmazingFeature'`)
4. Pushen Sie zum Branch (`git push origin feature/AmazingFeature`)
5. Öffnen Sie einen Pull Request

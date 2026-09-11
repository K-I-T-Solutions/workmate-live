# Workmate Live - Concept

**Ziel:** Open-Source Streamer.bot Alternative, die plattformunabhängig (Linux, macOS, Windows) läuft.

**Problem:** Streamer.bot ist ein mächtiges Tool für Stream-Automation, aber nur auf Windows verfügbar. Linux-Streamer haben keine vergleichbare Lösung mit GUI.

**Lösung:** Go-Backend + React Web-UI, die alle relevanten Streaming-Integrationen zentral steuert.

---

## Architektur

```
┌─────────────────────────────────────────────────┐
│                  Web UI (React)                  │
│         Dashboard / Commands / Settings          │
└──────────────────────┬──────────────────────────┘
                       │ WebSocket + REST API
┌──────────────────────┴──────────────────────────┐
│                Go Backend (Portal)               │
│  ┌──────────┐ ┌──────────┐ ┌──────────────────┐ │
│  │  Twitch   │ │   OBS    │ │     YouTube      │ │
│  │  Client   │ │  Client  │ │     Client       │ │
│  └──────────┘ └──────────┘ └──────────────────┘ │
│  ┌──────────┐ ┌──────────┐ ┌──────────────────┐ │
│  │  Command  │ │  Event   │ │      Auth        │ │
│  │  System   │ │  System  │ │   (JWT/Users)    │ │
│  └──────────┘ └──────────┘ └──────────────────┘ │
└──────────────────────┬──────────────────────────┘
                       │
┌──────────────────────┴──────────────────────────┐
│              Agent (Python, optional)             │
│           AI-gestützte Stream-Assistenz           │
└─────────────────────────────────────────────────┘
```

---

## Status: Was ist fertig?

### Infrastruktur
- [x] Go Backend mit Chi Router
- [x] JWT Authentication
- [x] WebSocket Hub (Echtzeit-Events an alle Clients)
- [x] YAML Config System mit Validation
- [x] React Frontend mit Vite + TypeScript + Tailwind
- [x] Dark Streaming Theme mit Sidebar-Navigation
- [x] Docker Support (Backend + Frontend)

### OBS Integration
- [x] OBS WebSocket v5 Client
- [x] Szenen auflisten & wechseln
- [x] Quellen auflisten & togglen
- [x] Streaming starten/stoppen
- [x] Recording starten/stoppen/pausieren
- [x] Echtzeit-Events via WebSocket

### Twitch Integration
- [x] Helix API Client (Stream Stats, Metadata Updates)
- [x] IRC Chat Client (Lesen + Schreiben)
- [x] EventSub WebSocket (Follows, Subs, Raids)
- [x] Chat Commands System (Built-in: ping, uptime, help, say)
- [x] Custom Commands CRUD mit YAML-Persistierung
- [x] Template-Variablen: `{user}`, `{channel}`, `{args}`, `{count}`
- [x] Command Cooldowns, Mod-Only, Enable/Disable

### YouTube Integration
- [x] Basic Client (Status, Stats)

### Agent (Python)
- [x] Separater Service mit REST API
- [x] Polling vom Portal

---

## Roadmap: Was fehlt noch?

### Phase 1: Event-Action System (Streamer.bot Kernfeature)
Das Herzstück von Streamer.bot: "Wenn X passiert, mache Y"

**Konzept:**
```yaml
triggers:
  - type: twitch_follow
    actions:
      - type: obs_scene_switch
        params: { scene: "Follow Alert" }
      - type: chat_message
        params: { message: "Danke für den Follow, {user}!" }
      - type: delay
        params: { seconds: 5 }
      - type: obs_scene_switch
        params: { scene: "Main" }
```

**Trigger-Typen:**
- Twitch: Follow, Sub, Raid, Bits, Chat-Keyword, Channel Points
- OBS: Szene gewechselt, Stream gestartet/gestoppt
- Timer: Intervall, Cron-Schedule
- Manual: Button in der UI

**Action-Typen:**
- OBS: Szene wechseln, Quelle togglen, Filter ändern
- Twitch Chat: Nachricht senden, Announcement, Timeout/Ban
- Sound: Audio abspielen
- HTTP: Webhook aufrufen
- Delay/Wait: Pause zwischen Actions
- Variable: Counter setzen/lesen

**Backend-Architektur:**
```
Event eingehend → Trigger Matcher → Action Queue → Action Runner
                                         ↓
                                   Sequentiell oder
                                   Parallel ausführen
```

**Dateien:**
- `internal/automation/trigger.go` — Trigger-Registry + Matcher
- `internal/automation/action.go` — Action-Registry + Runner
- `internal/automation/rule.go` — Rule Store (YAML-Persistierung)
- `internal/automation/engine.go` — Event Loop + Orchestrierung

### Phase 2: Timer & Scheduled Commands
- Wiederkehrende Chat-Nachrichten (z.B. "Folgt mir auf Twitter" alle 15 Min)
- Cron-basierte Aktionen
- Aktiv nur wenn Stream live ist (optional)

### Phase 3: Counters & Variablen
- Globale Variablen (z.B. Death Counter, !wins)
- Per-User Variablen (z.B. Punkte-System)
- Persistierung in YAML oder SQLite
- Nutzbar in Templates: `{var:death_count}`, `{user_var:points}`

### Phase 4: Sound Board
- Audio-Files abspielen als Action
- Sound Board UI mit Buttons
- Hotkey-Support (optional)

### Phase 5: Alerts & Overlays
- Browser Source Overlay für OBS
- Custom Alert Widgets (Follow, Sub, Raid)
- HTML/CSS/JS Templates
- WebSocket-basierte Echtzeit-Updates

### Phase 6: Erweiterte Twitch-Features
- Channel Points Integration
- Predictions erstellen/verwalten
- Polls erstellen
- Clips erstellen
- Mod-Actions (Timeout, Ban, Slow Mode)
- VIP Management

### Phase 7: Erweiterbarkeit
- Plugin-System (Go Plugins oder Script-basiert)
- Lua/JS Scripting für Custom Actions
- Community-Marketplace für Configs

---

## Technische Prinzipien

1. **Plattformunabhängig** — Go + Web UI, kein OS-Lock-in
2. **Single Binary** — Backend als einzelne ausführbare Datei
3. **Config as Code** — YAML-Dateien, versionierbar mit Git
4. **Echtzeit** — WebSocket für alle Live-Daten
5. **Modular** — Jede Integration ist ein eigenständiger Service
6. **Self-Hosted** — Läuft lokal, keine Cloud-Abhängigkeit
7. **Leichtgewichtig** — Minimale Resource-Nutzung

---

## Vergleich mit Streamer.bot

| Feature | Streamer.bot | Workmate Live |
|---------|-------------|---------------|
| Plattform | Windows only | Linux, macOS, Windows |
| UI | Desktop App (.NET) | Web UI (Browser) |
| OBS | OBS WebSocket | OBS WebSocket |
| Twitch | Vollständig | Chat, Events, API |
| YouTube | Teilweise | Basic |
| Chat Commands | Ja | Ja (Built-in + Custom) |
| Event-Actions | Ja (Kernfeature) | Geplant (Phase 1) |
| Timer | Ja | Geplant (Phase 2) |
| Counters | Ja | Geplant (Phase 3) |
| Sound Board | Ja | Geplant (Phase 4) |
| Overlays | Via OBS | Geplant (Phase 5) |
| Scripting | C# | Geplant (Phase 7) |
| Preis | Kostenlos | Open Source |
| Self-Hosted | Ja | Ja |

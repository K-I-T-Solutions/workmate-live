# Multi-Agent: Mehrere OBS-Rechner einzeln steuern

Entwurf, Stand 2026-09-14. Noch nicht umgesetzt — dieses Dokument beschreibt,
was gebaut werden müsste, in welcher Reihenfolge und was es kostet.

**Ziel:** Mehrere Rechner mit je eigener OBS-Instanz. Im Portal wird
ausgewählt, welcher gerade bedient wird. Jeder behält eigene Szenen, Quellen
und Streaming-Zustände.

```
                        ┌──────────── Portal ────────────┐
                        │  Hub: map[agentID]*Conn        │
                        │  aktiv im UI: cisco            │
                        └────┬──────────┬──────────┬─────┘
                   wss://    │          │          │
                   ┌─────────┘          │          └─────────┐
              ┌────┴────┐          ┌────┴────┐          ┌────┴────┐
              │  cisco  │          │  barry  │          │ laptop  │
              │  Studio │          │ Kamera 2│          │   IRL   │
              │  OBS ✓  │          │  OBS ✓  │          │  OBS ✓  │
              └─────────┘          └─────────┘          └─────────┘
```

## Ausgangslage

### Was bereits mehragentenfähig ist

Der Unterbau trägt das Konzept schon vollständig:

| Baustein | Datei | Zustand |
|---|---|---|
| Verbindungsregistry | `internal/agentlink/hub.go` | `map[string]*Conn`, beliebig viele Agents |
| Identität je Agent | `internal/agentlink/protocol.go` | `Hello{AgentID, Hostname, Version, OBSLocal, OBSActive}` |
| Kommando-Korrelation | `internal/agentlink/conn.go` | pro Verbindung, unabhängig |
| Statusspeicher | `internal/services/agent/cache.go` | `map[agentID]Status` |
| Auflistung | `GET /api/agents` | liefert alle |
| Gezielter Status | `GET /api/agent/status?agent_id=X` | funktioniert |
| Reconnect, Verdrängung | `hub.register/unregister` | pro Agent |

Ein zweiter Agent kann sich **heute schon** verbinden und erscheint korrekt in
der Liste. Das Protokoll ist agent-agnostisch.

### Was bei zwei Agents bricht

Drei Stellen verwerfen die Zuordnung. Heute unsichtbar, weil nur ein Agent
läuft — beim zweiten führt es zu stiller Fehlfunktion, nicht zu einem Fehler.

**1. OBS-Events verlieren den Absender.**
`cmd/workmate-live-portal/main.go:121` ruft `handleOBSEvent(obsEvent)` auf.
Die `agentID` liegt im Scope, wird aber nicht übergeben. Ein `scene_changed`
erreicht Frontend und Automation-Engine ohne Herkunft.

**2. Der Status-Broadcast ist mehrdeutig.**
`websocket.MessageTypeAgentStatus` transportiert `health.Status` — darin steht
`hostname`, aber keine Agent-ID. Zwei Agents überschreiben sich im Frontend
gegenseitig im Sekundentakt.

**3. Die Steuerung kennt nur einen Agent.**
`obs.RemoteController` fragt in jedem Aufruf `hub.OBSAgent()`. Alle elf
`/api/obs/*`-Routen landen bei genau diesem. Ein zweiter OBS-Rechner ist
sichtbar, aber nicht bedienbar. Ebenso hat `obs:scene_switch` in der
Automation kein Ziel-Feld.

Auch das Frontend hält je genau einen Zustand:
`obsStore.status: OBSStatus | null`, `agentStore.status: AgentStatus | null`.

## Zielbild

### Begriffe

- **Agent** — ein verbundener Rechner, identifiziert über `agent_id`
  (Standard: Hostname, überschreibbar in der Agent-Konfiguration).
- **Primärer Agent** — der ohne explizite Angabe angesprochene. Aus
  `agent.primary_id`, sonst der erste verbundene mit OBS. Bleibt wie heute.
- **Adressierung** — jede OBS-Operation nennt entweder einen Agent oder
  keinen; ohne Angabe gilt der primäre.

### Leitplanken

1. **Keine Breaking Changes.** Alle heutigen Aufrufe funktionieren weiter und
   sprechen den primären Agent an. Der laufende Betrieb mit `cisco` darf
   durch keine Stufe unterbrochen werden.
2. **Jede Stufe ist für sich lauffähig** und kann einzeln ausgerollt werden.
3. **Der Agent bleibt unverändert.** Das Link-Protokoll kennt bereits nur
   Punkt-zu-Punkt-Verbindungen — die Mehrfachverwaltung ist allein Sache des
   Portals. Kein neues Agent-Release nötig.

## Stufe 1 — Zuordnung reparieren

Behebt die drei Fehler oben. Sollte unabhängig von der Multi-Agent-Frage
passieren, weil es echte Bugs sind.

### Backend

**Events mit Herkunft** (`main.go`):

```go
// vorher
handleOBSEvent := func(event map[string]interface{}) { … }
handleOBSEvent(obsEvent)

// nachher
handleOBSEvent := func(agentID string, event map[string]interface{}) {
    hub.Broadcast(websocket.Message{
        Type: websocket.MessageTypeOBSEvent,
        Data: withAgent(agentID, event),   // ergänzt "agent_id"
    })

    vars := map[string]string{"agent": agentID}
    switch event["type"] {
    case "scene_changed":
        vars["scene"], _ = event["scene_name"].(string)
        autoEngine.Send(automation.Event{Type: "obs:scene_changed", Vars: vars})
    …
    }
}
handleOBSEvent(agentID, obsEvent)
```

Im Direktmodus (`obs.mode: direct`) gibt es keinen Agent — dort wird
`agent_id` leer gelassen bzw. auf `"local"` gesetzt.

**Status mit Kennung.** Neuer Umschlag statt nacktem `health.Status`:

```go
type AgentStatusMessage struct {
    AgentID string        `json:"agent_id"`
    Status  *agent.Status `json:"status"`
}
```

Das ist eine Formatänderung auf dem WebSocket. Damit das Frontend nicht
gleichzeitig umgestellt werden muss: beide Felder senden — `status` flach
weiterhin wie bisher, `agent_id` zusätzlich daneben. Das Frontend liest es,
sobald es soweit ist; danach kann die Doppelung entfallen.

**Neue Automation-Variable** `{agent}` in allen `obs:*`-Triggern.

### Aufwand

Klein. Drei Dateien im Backend, ein Feld im Frontend-Typ, Tests für beide
Event-Pfade (Link und Direktmodus).

## Stufe 2 — Steuerung adressierbar

### Routen

Kanonisch wird die Agent-Route, die bisherige bleibt als Kurzform für den
primären Agent:

```
GET  /api/agents/{agent}/obs/status
GET  /api/agents/{agent}/obs/scenes
POST /api/agents/{agent}/obs/scenes/switch
GET  /api/agents/{agent}/obs/sources?scene=…
POST /api/agents/{agent}/obs/sources/toggle
POST /api/agents/{agent}/obs/streaming/{start|stop}
POST /api/agents/{agent}/obs/recording/{start|stop|pause|resume}

GET  /api/obs/status            → primärer Agent (unverändert)
POST /api/obs/scenes/switch     → primärer Agent (unverändert)
…
```

Chi kann beides ohne Umbau:

```go
r.Route("/agents/{agent}/obs", func(r chi.Router) { … h.OBS.GetScenes … })
r.Route("/obs", func(r chi.Router) { … dieselben Handler … })
```

Der Handler liest `chi.URLParam(r, "agent")`; ist es leer, gilt der primäre.

### Controller

`obs.Controller` bleibt unverändert — das Interface beschreibt *eine*
OBS-Instanz. Neu kommt eine Auflösung davor:

```go
// ControllerFor liefert den Controller für einen bestimmten Agent.
// Leerer agentID bedeutet: primärer Agent.
type ControllerResolver interface {
    ControllerFor(agentID string) (Controller, error)
}
```

Zwei Implementierungen:

- `agentResolver` (Link-Betrieb) — baut pro Aufruf einen `RemoteController`
  auf die konkrete Verbindung. Existiert der Agent nicht oder ist er nicht
  verbunden: `ErrAgentNotFound` → HTTP 404 statt 500, damit die UI den
  Unterschied zwischen „Agent weg" und „OBS kaputt" zeigen kann.
- `singleResolver` (Direktmodus) — liefert immer denselben `*obs.Client` und
  lehnt eine abweichende `agent_id` ab.

`RemoteController` bekommt dafür einen zweiten Konstruktor, der eine feste
Verbindung statt der Hub-Auswahl nutzt. Die bestehende Variante bleibt.

### Handler

Die elf Handler ändern sich nach demselben Muster:

```go
func (h *OBSHandler) GetScenes(w http.ResponseWriter, r *http.Request) {
    ctrl, err := h.resolve(r)          // neu: eine Zeile
    if err != nil { writeAgentError(w, err); return }

    scenes, err := ctrl.GetScenes()    // Rest unverändert
    …
}
```

### Aufwand

Mittel. Ein neues Interface, zwei Resolver, elf Handler um je zwei Zeilen
ergänzt, Routen verdoppelt, Tests für Auflösung und Fehlerfälle.

## Stufe 3 — Automation mit Ziel

### Actions

`params.agent` wird optional ergänzt:

```yaml
- name: Bei Raid auf Alert-Szene
  trigger:
    type: twitch:raid
  actions:
    - type: obs:scene_switch
      params:
        agent: cisco        # leer = primärer Agent
        scene: Raid
    - type: obs:scene_switch
      params:
        agent: barry
        scene: Kamera 2
```

### Trigger filtern

Damit eine Regel nur auf einen bestimmten Rechner reagiert:

```yaml
  trigger:
    type: obs:scene_changed
    agent: barry            # leer = jeder Agent
```

Die Engine vergleicht `trigger.agent` mit `event.Vars["agent"]` aus Stufe 1.

### Engine

`automation.Engine` hält heute genau einen `OBSExecutor` (`engine.go:19`).
Das wird ein Resolver:

```go
type OBSResolver interface {
    ExecutorFor(agentID string) (OBSExecutor, error)
}
```

`executeSceneSwitch` und `executeSourceToggle` lösen dann pro Aufruf auf.
Fehlt der Agent, schlägt die Aktion mit klarer Meldung fehl, statt still ins
Leere zu laufen — wie heute schon bei fehlender OBS-Verbindung.

### Optional: Broadcast

`agent: "*"` schaltet alle verbundenen OBS-Instanzen gleichzeitig. Billig zu
bauen (Schleife über `hub.List()`), aber eine eigene Fehlersemantik: Was
gilt, wenn drei von vier Agents antworten? Vorschlag: Teilerfolg ist
Erfolg, fehlgeschlagene Agents werden geloggt und im Activity-Log genannt.
Erst bauen, wenn ein konkreter Anwendungsfall da ist.

### Aufwand

Mittel. Betrifft `types.go`, `engine.go`, `obs.go` in `internal/automation`,
dazu das Regel-Schema und dessen Validierung.

## Stufe 4 — Frontend

### Stores

`obsStore` hält heute einen Zustand. Künftig einen je Agent plus Auswahl:

```ts
interface OBSStore {
  selected: string | null                 // aktive agent_id
  byAgent: Record<string, {
    status: OBSStatus | null
    scenes: Scene[]
    sources: Source[]
  }>
  select: (agentID: string) => void
  setStatus: (agentID: string, status: OBSStatus) => void
  …
}
```

Analog `agentStore`: `byAgent: Record<string, AgentStatus>`.

Die WebSocket-Behandlung (`services/websocket.ts:72,99`) schreibt dann in den
Eintrag der mitgelieferten `agent_id` statt in einen globalen Slot.

### Oberfläche

- **Agent-Umschalter** oben auf der OBS-Seite: Segmented Control bei zwei bis
  drei Rechnern, Dropdown ab vier. Zeigt je Agent den Verbindungszustand
  (verbunden / OBS getrennt / offline) über denselben Punkt wie heute.
- **Auswahl merken** in `localStorage`, damit ein Reload nicht auf den
  primären Agent zurückspringt.
- **Dashboard**: eine Karte je Agent statt einer globalen Systemkarte.
  Bei einem einzelnen Agent sieht es aus wie heute — der Umschalter wird
  ausgeblendet, solange nur einer verbunden ist.
- **Automation-Editor**: Agent-Auswahl in den OBS-Actions, vorbelegt mit
  „primärer Agent".

### Aufwand

Der größte Einzelposten. Zwei Stores, die WebSocket-Behandlung, OBS-Seite,
Dashboard und der Automation-Editor.

## Was sich nicht ändert

- **Der Agent.** Kein neues Release nötig, keine Konfigurationsänderung.
- **Das Link-Protokoll.** `Frame`, `Hello`, alle Methoden bleiben.
- **Der Direktmodus.** `obs.mode: direct` bleibt unverändert nutzbar und
  verhält sich wie ein einzelner, fest verdrahteter Agent.
- **Bestehende Automation-Regeln.** Ohne `agent`-Feld laufen sie wie bisher
  gegen den primären Agent.

## Reihenfolge und Risiko

| Stufe | Nutzen | Risiko | Rückwärtskompatibel |
|---|---|---|---|
| 1 — Zuordnung | behebt echte Bugs | gering | ja, Frontend liest neue Felder optional |
| 2 — API | zweiter Rechner steuerbar | gering | ja, alte Routen bleiben |
| 3 — Automation | Regeln je Rechner | mittel | ja, `agent` ist optional |
| 4 — Frontend | bedienbar statt nur API | gering, aber viel Fläche | ja |

Stufe 1 ist die einzige, die unabhängig vom Multi-Agent-Ziel Sinn ergibt.
Stufe 2 macht das System technisch mehragentenfähig — ab da ist ein zweiter
OBS-Rechner über die API vollständig bedienbar, auch ohne Stufe 4. Damit
lässt sich der Nutzen prüfen, bevor die Oberfläche umgebaut wird.

## Offene Entscheidungen

1. **Agent-Kennung in der UI.** `agent_id` (technisch, stabil) oder
   `hostname` (lesbarer, kann kollidieren)? Vorschlag: `hostname` anzeigen,
   `agent_id` als Tooltip, adressiert wird immer über `agent_id`.

2. **Verhalten bei abwesendem Agent.** 404 mit klarer Meldung (Vorschlag)
   oder stiller Fallback auf den primären? Letzteres kann dazu führen, dass
   versehentlich der falsche Rechner geschaltet wird — deshalb 404.

3. **Braucht ein Agent ein Anzeigelabel?** Ein optionales `display_name` in
   der Agent-Konfiguration (`"Studio"`, `"Kamera 2"`) wäre in der UI
   deutlich lesbarer als Hostnamen. Kleine Protokollergänzung im `Hello`.

4. **Statushistorie.** Aktuell hält der `StatusCache` nur den letzten Stand je
   Agent. Für ein Dashboard mit Verlauf (CPU, Frames) bräuchte es einen
   Ringpuffer — eigenes Thema, nicht Teil dieses Entwurfs.

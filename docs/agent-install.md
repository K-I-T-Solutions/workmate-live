# Agent installieren

Der Agent läuft auf **Linux, Windows und macOS**. Er überwacht das System und
steuert im Auftrag des Portals die lokale OBS-Instanz. Die Verbindung baut er
selbst nach außen auf — der Rechner braucht keinen eingehenden Port und
funktioniert hinter NAT, Fritzbox oder Firewall.

## Vorbereitung: OBS

In OBS unter **Werkzeuge → WebSocket-Servereinstellungen**:

- WebSocket-Server aktivieren
- Port notieren (Standard `4455`)
- Passwort anzeigen lassen und notieren

Der Server darf auf `localhost` gebunden bleiben. Der Agent läuft auf demselben
Rechner und erreicht ihn dort — nach außen wird nichts geöffnet.

## Binary holen

Fertige Binaries liegen unter
[Releases](https://github.com/K-I-T-Solutions/workmate-live/releases),
gebaut für:

| Plattform | Datei |
|---|---|
| Windows (Intel/AMD) | `workmate-live-agent_windows_amd64.zip` |
| Windows (ARM) | `workmate-live-agent_windows_arm64.zip` |
| macOS (Apple Silicon) | `workmate-live-agent_darwin_arm64.tar.gz` |
| macOS (Intel) | `workmate-live-agent_darwin_amd64.tar.gz` |
| Linux (x86-64) | `workmate-live-agent_linux_amd64.tar.gz` |
| Linux (ARM) | `workmate-live-agent_linux_arm64.tar.gz` |

Selbst bauen geht überall gleich:

```bash
cd agent
go build -o workmate-live-agent ./cmd/workmate-live-agent
```

## Konfiguration

Auf allen Plattformen identisch. `config.example.yaml` nach `config.yaml`
kopieren und anpassen:

```yaml
obs:
  enabled: true
  host: "127.0.0.1"
  port: 4455
  password: "<Passwort aus OBS>"
  reconnect_delay: 5s

portal:
  enabled: true
  url: "https://live.workmate.kit-it-koblenz.de"
  api_key: "<derselbe Wert wie agent.api_key im Portal>"
  agent_id: ""      # leer = Hostname des Rechners
  timeout: 10s
  retry_attempts: 3
  retry_delay: 5s
```

Laufen mehrere Rechner mit, vergib je einen eigenen `agent_id` — sonst
verdrängen sie sich gegenseitig im Portal.

Gesucht wird die Datei hier:

| | |
|---|---|
| alle | `./config.yaml` |
| Linux | `~/.config/workmate-agent/config.yaml`, `/etc/workmate-agent/config.yaml` |
| macOS | `~/Library/Application Support/workmate-agent/config.yaml`, `~/.config/…`, `/etc/…` |
| Windows | `%AppData%\workmate-agent\config.yaml`, `%USERPROFILE%\.config\workmate-agent\config.yaml` |

Oder direkt angeben: `workmate-live-agent --config <Pfad>`

---

## Windows

Archiv entpacken, etwa nach `C:\Program Files\Workmate Live Agent\`.
Zum Testen in der PowerShell:

```powershell
cd "C:\Program Files\Workmate Live Agent"
.\workmate-live-agent.exe --config .\config.yaml
```

Erwartete Ausgabe:

```
obs control enabled for 127.0.0.1:4455
link: connecting to portal at https://live.workmate.kit-it-koblenz.de
obs: connected to 127.0.0.1:4455
link: connected to portal
```

### Als Dienst einrichten

Windows bringt keinen Dienst-Wrapper für beliebige Programme mit. Zwei Wege:

**Aufgabenplanung** (ohne Zusatzsoftware) — als Administrator:

```powershell
$exe = "C:\Program Files\Workmate Live Agent\workmate-live-agent.exe"
$cfg = "C:\Program Files\Workmate Live Agent\config.yaml"

$action  = New-ScheduledTaskAction -Execute $exe -Argument "--config `"$cfg`"" `
           -WorkingDirectory (Split-Path $exe)
$trigger = New-ScheduledTaskTrigger -AtLogOn
$settings = New-ScheduledTaskSettingsSet -AllowStartIfOnBatteries `
            -DontStopIfGoingOnBatteries -RestartCount 999 -RestartInterval (New-TimeSpan -Minutes 1)

Register-ScheduledTask -TaskName "Workmate Live Agent" `
  -Action $action -Trigger $trigger -Settings $settings -RunLevel Highest
```

Beim Anmelden startet der Agent dann automatisch. `-AtLogOn` ist hier richtig:
OBS läuft ohnehin nur in einer angemeldeten Sitzung.

**NSSM** (komfortabler, echter Dienst) — falls installiert:

```powershell
nssm install WorkmateLiveAgent "C:\Program Files\Workmate Live Agent\workmate-live-agent.exe" --config config.yaml
nssm set WorkmateLiveAgent AppDirectory "C:\Program Files\Workmate Live Agent"
nssm start WorkmateLiveAgent
```

### Windows-Defender

Beim ersten Start kann SmartScreen anschlagen, weil das Binary nicht signiert
ist: **Weitere Informationen → Trotzdem ausführen**. Der Agent öffnet keinen
Port nach außen, braucht also keine Firewall-Freigabe.

---

## macOS

```bash
tar -xzf workmate-live-agent_darwin_arm64.tar.gz
xattr -d com.apple.quarantine workmate-live-agent   # Gatekeeper: nicht signiert
./workmate-live-agent --config ./config.yaml
```

Ohne den `xattr`-Aufruf meldet macOS „kann nicht geöffnet werden, da der
Entwickler nicht verifiziert werden kann".

### Als Dienst einrichten

`~/Library/LaunchAgents/de.kit.workmate-live-agent.plist` anlegen:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>de.kit.workmate-live-agent</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/workmate-live-agent</string>
        <string>--config</string>
        <string>/Users/DEINNAME/Library/Application Support/workmate-agent/config.yaml</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>/tmp/workmate-live-agent.log</string>
    <key>StandardErrorPath</key>
    <string>/tmp/workmate-live-agent.log</string>
</dict>
</plist>
```

Laden und prüfen:

```bash
launchctl load ~/Library/LaunchAgents/de.kit.workmate-live-agent.plist
tail -f /tmp/workmate-live-agent.log
```

Ein **LaunchAgent** (nicht LaunchDaemon) ist hier richtig: Der Agent braucht
die Benutzersitzung, in der auch OBS läuft.

### Berechtigungen

Für die Kameraliste fragt macOS unter Umständen nach der Kamera-Berechtigung.
Wird sie verweigert, bleibt nur diese Liste leer — OBS-Steuerung und alles
Übrige funktionieren weiter.

---

## Linux

```bash
tar -xzf workmate-live-agent_linux_amd64.tar.gz
sudo install -m 755 workmate-live-agent /usr/local/bin/
./workmate-live-agent --config ./config.yaml
```

systemd-Unit nach `/etc/systemd/system/workmate-live-agent.service`:

```ini
[Unit]
Description=Workmate Live Agent
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=DEINNAME
ExecStart=/usr/local/bin/workmate-live-agent --config /home/DEINNAME/.config/workmate-agent/config.yaml
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl enable --now workmate-live-agent
journalctl -u workmate-live-agent -f
```

---

## Prüfen

Im Portal unter **Einstellungen → Agent** erscheint der Rechner, oder per API:

```bash
curl -s -H "Authorization: Bearer $TOKEN" \
  https://live.workmate.kit-it-koblenz.de/api/agents | jq
```

Erwartet: `"connected": true` und `"obs_active": true`, sobald OBS läuft.

Startet OBS erst später, verbindet der Agent innerhalb von fünf Sekunden von
selbst nach — dasselbe gilt nach einem Neustart des Portals.

## Was sich zwischen den Plattformen unterscheidet

OBS-Steuerung und Portal-Verbindung sind überall gleich. Nur die
Systemüberwachung greift auf das zu, was das jeweilige System anbietet:

| Prüfung | Linux | Windows | macOS |
|---|---|---|---|
| GPU | `/dev/dri`, sysfs | Registry | `system_profiler` |
| Audio | PipeWire-Socket | WASAPI-Endpunkte | `coreaudiod` |
| Kameras | `/dev/video*` | Registry | `system_profiler` |
| OBS-Prozess | `/proc` | Prozessliste (Win32) | `pgrep` |

Auf macOS kosten die `system_profiler`-Aufrufe für GPU und Kameras rund eine
Sekunde. Sie werden deshalb einmalig beim Start gelesen — eine später
angesteckte Kamera erscheint erst nach einem Neustart des Agents. Alles
andere wird weiterhin laufend aktualisiert.

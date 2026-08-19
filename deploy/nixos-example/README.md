# NixOS deployment example

Deployment is a static binary + a systemd unit + a reverse proxy. Nothing here is
imported by the build — these are copy-paste examples.

`event-sensor.nix` targets a fresh-services module: personal convention, not a
universal NixOS module, and inert without it. `systemd-unit.example` is the portable
one — plain systemd, any Linux, no nix.

| File                       | Purpose                                                        |
| -------------------------- | -------------------------------------------------------------- |
| `event-sensor.nix`         | service definition; the unit and proxy config generate from it |
| `event-sensor-env.example` | secrets for systemd `EnvironmentFile` — never in the nix store |
| `systemd-unit.example`     | plain systemd unit, for machines without the module            |

Port 6034 is arbitrary (next free).

## Steps

1. Build and place the binary:

   ```sh
   just build
   install -Dm755 event-sensor /home/username/custom-systemd/event-sensor/event-sensor-bin
   ```

2. Generate a secret and keep it out of the nix store:

   ```sh
   openssl rand -hex 32   # JWT_SECRET
   ```

   Put it in an agenix/sops file decrypted to `/run/agenix/event-sensor-env`, then point
   the unit at it:

   ```sh
   systemctl edit event-sensor
     [Service]
     EnvironmentFile=/run/agenix/event-sensor-env
   ```

3. Copy `event-sensor.nix` into the module's services directory and import it.

## Break-glass (forgot the password)

```sh
sqlite3 /home/username/0/apps-media/event-sensor/event-sensor.db \
  "UPDATE users SET password='$(htpasswd -bnBC 10 "" 'newpass' | tr -d ':\n')' WHERE username='admin';"
```

The app binds 127.0.0.1; off-loopback binds refuse to start with the default
`JWT_SECRET` (see `internal/config`).

# After a host reboot: agentbus, herdr-hosted agents, decide, slackwatch, and a hub-mode client

Run these in order after the host comes back. Each command is followed
by the one-line output that means it is good.

## User units

On Linux, nothing in the user session should have failed, and the two
units this stack uses should be active. `agentbus install` writes
`agentbus.service`. slackwatch's unit is
[contrib/slackwatch.service](../slackwatch/contrib/slackwatch.service).

```bash
systemctl --user --failed
```

Good: `0 loaded units listed.`

```bash
systemctl --user is-active agentbus.service slackwatch.service
```

Good: two lines, both `active`.

On macOS the supervisor is launchd, not systemd. The label in this
repo is `com.agentbus.agentbusd`.

```bash
launchctl list | grep agentbus
```

Good: a line for `com.agentbus.agentbusd` whose first column is a pid,
not `-`. Other `com.agentbus.*` lines may appear too.

## agentbus versus herdr start order

```bash
agentbus search
```

Good: one block per reachable agent. Empty output is the failure below.

```bash
agentbus list
```

Good: herdr-registered names are not all `liveness: unreachable`.

If `agentbus search` printed nothing, or every herdr-registered name
is `liveness: unreachable`, right after boot, the daemon started
before the herdr server. Since
[#767](https://github.com/jaeyeom/experimental/issues/767) it keeps
retrying and recovers by itself once herdr is up, usually within a
minute. Held messages then drain on their own. A daemon older than
#767 does not self-recover, so it still needs the restart below.

```bash
agentbus doctor
```

Good: the output does not contain `herdr transport unreachable`.

Restart only if `agentbus doctor` still prints
`herdr transport unreachable` after about a minute. Wait until the
log is quiet first. Skip the restart when the checks above are
already good.

```bash
agentbus log --since 10s
```

Good: no lines. If events are still printing, wait and run it again.

```bash
systemctl --user restart agentbus
```

Good: no output and exit 0, then `systemctl --user is-active agentbus.service` prints `active`.

On macOS, restart the same launchd job. `launchctl kickstart` without
`-k` does not replace a job that is already running.

```bash
launchctl kickstart -k "gui/$(id -u)/com.agentbus.agentbusd"
```

Good: no output and exit 0, and `com.agentbus.agentbusd` has a pid in `launchctl list`.

## herdr tabs

Tabs, panes, and agent sessions are restored, and agents resume with
their session ids.

```bash
herdr agent list
```

Good: each restored agent is listed with `agent_session.value` set.

Read any tab whose last turn ended on a usage-limit or error and
resume it. Text the user typed but did not send, still in an input
box, is the user's to send, never the agent's.

## Scheduled agent jobs

Cron-like schedules registered inside agent sessions survived the
reboot. They are listed across sessions, not only in the tab that
created them.

Good: each coordinating role's job shows up in that schedule list.

Two gotchas:

- Recurring jobs expire on their own after a fixed lifetime (currently
  7 days). Check each coordinating role's job and re-create any near
  expiry.
  Good: nothing you still need is near that 7-day expiry.
- A job can stay listed while its tab missed runs, because the turn
  died on a usage-limit or error. Check the tab's last run time, not
  only that the job is listed.
  Good: the last run is a turn that finished, not a usage-limit or error.

Temporary scratch directories under `/tmp` are wiped. Tooling that
must survive restarts lives under ~~/.local/state~ (or
`$XDG_STATE_HOME` when that is set), not in a session scratchpad.

## Hub-mode client

On the client, the tunnel reconnects on its own through the service
manager: `agentbus-tunnel.service` on Linux, `com.agentbus.tunnel` on
macOS.

```bash
agentbus doctor
```

Good: the tunnel socket line ends in `ok`, and the ssh line is `ssh <hub>: ok (its ssh port answers)`.

When the tunnel socket is up, an older exit 255 is labelled stale:
the line says the tunnel is up and that the old exit is stale. When
the tunnel is down, doctor names the missing piece
(`ssh_identity_agent` unset, ~~/.ssh/agent.sock~ missing, or the
agent has no keys) and one fix command. That leftover does not by
itself mean the tunnel is still down. Confirm the tunnel unit has a
running pid, and that `agentbus list` works from this client.

```bash
systemctl --user show -p MainPID --value agentbus-tunnel.service
```

Good: a pid other than `0` (Linux).

```bash
launchctl list | grep com.agentbus.tunnel
```

Good: a numeric pid in the first column, not `-` (macOS).

```bash
agentbus list
```

Good: names print, and the command does not fail with an ssh authentication error.

If ssh to the hub needs a passphrase-protected key, the stable agent
socket has to be in place. Without it, non-login shells cannot reach
the hub. See [#740](https://github.com/jaeyeom/experimental/issues/740)
and [#741](https://github.com/jaeyeom/experimental/issues/741). The
line already under [Hub mode](README.md#hub-mode-a-second-machine)
is

```toml
ssh_identity_agent = "~/.ssh/agent.sock"
```

## Final check

On every host:

```bash
decide list
```

Good: exit 0, with one line per open item, or no lines when the inbox is empty.

```bash
agentbus list
```

Good: the names you expect show `liveness: alive`.

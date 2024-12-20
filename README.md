# unlockproxy

[![ci](https://img.shields.io/github/actions/workflow/status/cetteup/unlockproxy/ci.yaml?label=ci)](https://github.com/cetteup/unlockproxy/actions?query=workflow%3Aci)
[![Go Report Card](https://goreportcard.com/badge/github.com/cetteup/unlockproxy)](https://goreportcard.com/report/github.com/cetteup/unlockproxy)
[![License](https://img.shields.io/github/license/cetteup/unlockproxy)](/LICENSE)
[![Last commit](https://img.shields.io/github/last-commit/cetteup/unlockproxy)](https://github.com/cetteup/unlockproxy/commits/main)

A simple HTTP proxy that unlocks all BF2 weapons for every player

## Usage

### Running unlockproxy

You can run unlockproxy on Windows or Linux, either directly or via Docker. Please note that unlockproxy usually cannot run on the same host as the game server.

On any Linux using systemd, you can run unlockproxy as a simple service.

```ini
[Unit]
Description=Battlefield 2 unlock proxy
After=network-online.target

[Install]
WantedBy=multi-user.target

[Service]
Type=simple

WorkingDirectory=/opt/unlockproxy
ExecStart=/opt/unlockproxy/unlockproxy

Restart=always
RestartSec=1m

StandardOutput=syslog
StandardError=syslog
SyslogIdentifier=%n

User=unlockproxy
```

### Redirecting ASPX HTTP traffic to unlockproxy

To make your Battlefield 2 server use unlockproxy, you need to redirect all HTTP traffic for the game's ASPX endpoints to unlockproxy. On Linux, you can use `iptables` to achieve this. Using the IP addresses behind `servers.bf2hub.com` as an example, you could use (as root/with `sudo`):

```sh
iptables -t nat -A OUTPUT -p tcp -d 92.51.180.45,92.51.181.102 --dport 80 -j DNAT --to-destination 192.168.100.50:8080
```

This rule will redirect any HTTP traffic designated to BF2Hub's servers to an instance of unlockproxy running on `192.168.100.50`. Make sure you have a setup to persist `iptables` rules, otherwise they will lost on reboot.

## How it works

When a players joins a Battlefield 2 server using global unlocks (`sv.useGlobalUnlocks 1`), the server makes an HTTP request to `getunlocksinfo.aspx`. The response to that HTTP call tells the server which weapon unlocks it should enable for the player. It lists an `id` and a `state` for each unlocked weapon. As an example, this is a response from BF2Hub for a player who has unlocked all weapons.

```
O
H	pid	nick	asof
D	500362798	mister249	1676835714
H	enlisted	officer
D	0	0
H	id	state
D	11	s
D	22	s
D	33	s
D	44	s
D	55	s
D	66	s
D	77	s
D	88	s
D	99	s
D	111	s
D	222	s
D	333	s
D	444	s
D	555	s
$	130	$
```

Usually, this response will of course be different for each player. That's where unlockproxy comes in. It always returns the full list of weapons as unlocked when `getunlocksinfo.aspx` is called. Thus any player joining a server using unlockproxy will be able to use all unlocks. Any other ASPX HTTP calls made by the server (e.g. `VerifyPlayer.aspx`) are simply forwarded to a "real" Battlefield 2 stats backend (e.g. BF2Hub).

![image](https://github.com/user-attachments/assets/85c6ef53-b8d5-40b2-b593-8cd3421133aa)

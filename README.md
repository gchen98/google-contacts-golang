# Google Contacts Microservice (Go)

A small REST API, written in Go with [Gin](https://github.com/gin-gonic/gin), that lets people on the intranet list, search, add, update, and delete Google Contacts for any user in a Google Workspace domain. It needs no per-user OAuth flow.

It authenticates as a **Google service account with domain-wide delegation**. For each request, it impersonates the Workspace user named in the URL and calls the [Google People API](https://developers.google.com/people) on that user's behalf.

```
 intranet client ──HTTP/JSON──▶ contacts_go (Gin, :8340) ──JWT as <username>──▶ Google People API
```

---

## Features

| Method | Endpoint | Body | Description |
|--------|----------|------|-------------|
| `GET`  | `/ping` | — | Health check. Returns `{"message":"pong"}`. |
| `GET`  | `/ListContacts/:username` | — | Returns all of the user's contacts. Pages through the API 500 at a time. |
| `GET`  | `/SearchContacts/:username?query=<term>` | — | Searches the user's contacts by name, email, phone, etc. |
| `POST` | `/AddContacts/:username` | JSON array of [`Person`](https://developers.google.com/people/api/rest/v1/people#Person) | Creates contacts in one batch. |
| `POST` | `/UpdateContacts/:username` | JSON array of `Person` (with `resourceName` and `etag`) | Updates contacts in one batch. |
| `POST` | `/DeleteContacts/:username` | JSON array of resource names | Deletes contacts in one batch. |

`:username` is the full Workspace email address of the user whose contacts you want to work with, such as `alice@example.com`.

The service reads and writes these contact fields:

```
names, emailAddresses, phoneNumbers, organizations, addresses, biographies, userDefined
```

Responses also include `metadata`.

---

## Usage examples

```bash
HOST=http://localhost:8340
USER=alice@example.com

# Health check
curl $HOST/ping

# List all contacts
curl $HOST/ListContacts/$USER

# Search
curl "$HOST/SearchContacts/$USER?query=smith"

# Add contacts
curl -X POST -H 'Content-Type: application/json' \
     -d @backend/add_contacts.json  $HOST/AddContacts/$USER

# Update contacts
curl -X POST -H 'Content-Type: application/json' \
     -d @backend/update_contacts.json  $HOST/UpdateContacts/$USER

# Delete contacts
curl -X POST -H 'Content-Type: application/json' \
     -d '["people/c4896693973121291486"]'  $HOST/DeleteContacts/$USER
```

There are sample request bodies in `backend/` (`add_contacts.json`, `update_contacts.json`, `delete.json`). `contacts.json` shows example output from a list call.

### Notes on payloads

- **Add:** if you start from contacts returned by a list or search call, remove `resourceName`, `etag`, and every `metadata.source.id` first. Google assigns these values itself.
- **Update:** each `Person` must include its `resourceName` (`people/c…`) and its current `etag`. Get both from a fresh list or search call. If the `etag` is stale, Google rejects the update.
- **Delete:** the body is a plain JSON array of `resourceName` strings.
- **Search:** Google's `searchContacts` uses a lazily built cache. The first search for a user can return empty results until that cache warms up.

---

## Project layout

```
.
├── main.go                  # Entry point: calls api.StartServer()
├── go.work                  # Go workspace tying the modules together
├── api/
│   └── webservice.go        # Gin router and HTTP handlers
├── backend/
│   ├── people_api.go        # Service-account auth + People API calls
│   ├── ldap.go              # Experimental LDAP lookup (not wired into the API)
│   ├── config -> ../config  # Symlink so `go test` in backend/ finds the secret
│   └── *.json               # Sample request payloads
├── config/
│   ├── client_secret.json   # Service-account key (NOT in git)
│   └── contacts-go.service  # systemd unit file
└── greetings/               # Go tutorial module (unused by the service)
```

---

## Prerequisites

### Go

Go **1.25.4** or newer.

### Google Workspace service account

1. In the [Google Cloud Console](https://console.cloud.google.com/), create a project (or pick an existing one) and **enable the People API**.
2. Create a **service account** and download a **JSON key** for it.
3. Turn on **domain-wide delegation** for the service account and note its **Client ID**.
4. In the Google Workspace **Admin Console**, go to **Security → Access and data control → API controls → Manage Domain Wide Delegation**. Add the Client ID with this scope:
   ```
   https://www.google.com/m8/feeds
   ```
5. Save the downloaded key as `config/client_secret.json`. The file is listed in `.gitignore`. **Never commit it.**

---

## Building

```bash
git clone <repo-url> google-contacts-golang
cd google-contacts-golang

go build -o contacts_go .
```

This produces a single static binary, `contacts_go`, in the project root.

To run it locally:

```bash
./contacts_go
# [GIN-debug] Listening and serving HTTP on 127.0.0.1:8340
```

> **Important:** the service looks for its key at the **relative** path `config/client_secret.json`. Always run it from a directory that contains `config/client_secret.json`. Under systemd, that directory is `WorkingDirectory`.

For production, you can turn off Gin's debug output:

```bash
export GIN_MODE=release
```

---

## Deploying with systemd

The repo includes a unit file at `config/contacts-go.service`:

```ini
[Unit]
Description=Google Contacts Microservice
After=network.target

[Service]
Type=simple
User=garyc
WorkingDirectory=/home/garyc/code/google-contacts/contacts_go
ExecStart=/home/garyc/code/google-contacts/contacts_go/contacts_go
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

### 1. Build and stage the deployment directory

Before you start, change `User`, `WorkingDirectory`, and `ExecStart` in the unit file to match your host. The commands below use the paths from the unit file as shipped:

```bash
DEPLOY=/home/garyc/code/google-contacts/contacts_go

go build -o contacts_go .

mkdir -p $DEPLOY/config
cp contacts_go $DEPLOY/
cp config/client_secret.json $DEPLOY/config/
chmod 600 $DEPLOY/config/client_secret.json
```

The deployment directory should then look like this:

```
$DEPLOY/
├── contacts_go
└── config/
    └── client_secret.json
```

### 2. Install and start the service

```bash
sudo cp config/contacts-go.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now contacts-go.service
```

### 3. Verify

```bash
systemctl status contacts-go.service
curl http://127.0.0.1:8340/ping        # → {"message":"pong"}
journalctl -u contacts-go.service -f   # follow logs
```

### Upgrading

```bash
go build -o contacts_go .
cp contacts_go $DEPLOY/
sudo systemctl restart contacts-go.service
```

If you edit the unit file, run `sudo systemctl daemon-reload` before you restart.

---

## Network access and security

- **Bind address:** the server listens on `127.0.0.1:8340` (set in `api/webservice.go`), so only the local host can reach it. To serve the rest of the intranet, do one of the following:
  - put a reverse proxy (nginx, Caddy, Apache) in front of it, which also lets you add TLS and authentication, **or**
  - change the bind address in `r.Run(...)` to `0.0.0.0:8340` or a specific LAN interface, then rebuild.
- **No built-in authentication:** anyone who can reach the port can read and change the contacts of **any user in the domain**, because the service account impersonates whichever `:username` the caller supplies. Keep the service on a trusted network segment, or put authentication in front of it.
- **Protect the key:** `client_secret.json` grants domain-wide access. Keep it `chmod 600`, readable only by the service user.

---

## Testing

The backend tests run against the **live** Google APIs and use real accounts, so they need a valid `config/client_secret.json`. The contact tests (`TestAddContacts`, `TestUpdateContacts`, `TestDeleteContacts`) start with `t.Skip(...)`. Remove that line from the test you want, then run it on its own with `-run`:

```bash
cd backend
go test -run TestUpdateContacts -v
```



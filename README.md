[![Go Report Card](https://goreportcard.com/badge/github.com/MonsieurTib/service-bus-tui)](https://goreportcard.com/report/github.com/MonsieurTib/service-bus-tui)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=MonsieurTib_service-bus-tui&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=MonsieurTib_service-bus-tui)

# service-bus-tui

A terminal-based explorer for Azure Service Bus.

> **Work in Progress**: This project is currently under active development.

## Preview

<img width="2097" height="1207" alt="asb-tui" src="https://github.com/user-attachments/assets/ed282136-e839-46cc-bd6a-51a274e9372f" />

## Features

### Authentication
- Azure CLI authentication (uses existing `az login` session)
- Interactive browser authentication
- Service principal (client ID / client secret)
- Connection string
- Emulator (one-click connect to local Service Bus emulator)

### Namespace Discovery
- Automatically lists all Service Bus namespaces across your Azure subscriptions
- Displays namespace name, subscription ID, and resource group

### Resource Browsing
- Tree-based navigation of namespaces
- List topics and queues
- Expand topics to view subscriptions
- Expand queues to view active and DLQ messages
- View active messages and dead-letter queue (DLQ) messages per subscription or queue

### Message Viewing
- Peek messages from subscriptions and queues (active and DLQ)
- Tabular display with sequence number, message ID, subject, enqueued time, and body preview
- JSON body formatting in preview
- Paginated browsing (100 messages per page) with automatic page detection at boundaries

### Message Resending
- Resend selected messages back to their source topic or queue
- Works on both active and dead-letter messages (DLQ replay)
- Multi-message selection support
- Option to keep original or generate new Message IDs
- Real-time progress tracking

### Navigation
- Keyboard-driven interface
- `up/down` or `j/k`: Navigate items
- `left/right` or `h/l` or `enter`: Expand/collapse nodes
- `tab`: Switch between namespace tree and messages pane
- `ctrl+n` / `ctrl+p`: Next/previous page of messages
- `space`: Select/deselect messages
- `R`: Resend selected messages
- `esc`: Go back
- `ctrl+c`: Quit

## Installation

### Homebrew (macOS/Linux)

```bash
brew tap monsieurtib/tap
brew install service-bus-tui
```

### Winget (Windows)

```powershell
winget install MonsieurTib.service-bus-tui
```

### Go Install

```bash
go install github.com/MonsieurTib/service-bus-tui@latest
```

### Build from source

```bash
git clone https://github.com/MonsieurTib/service-bus-tui.git
cd service-bus-tui
go build -o service-bus-tui .
```

## Usage

```bash
service-bus-tui
```

Select an authentication method, choose a namespace, and browse your Service Bus resources.

## Emulator Support

Supports the [Azure Service Bus emulator](https://learn.microsoft.com/en-us/azure/service-bus-messaging/overview-emulator) running in Docker. Select **Emulator (localhost)** from the auth menu for default port.

**Limitation**: Total message count is unavailable (SDK bug) — page indicator shows "?" but pagination works normally.

## Requirements

- Go 1.24+
- Azure subscription with Service Bus namespaces (or the local emulator)
- For Azure CLI auth: `az login` must be completed beforehand
- For Emulator: Docker with the [Service Bus emulator](https://github.com/Azure/azure-service-bus-emulator-installer) running

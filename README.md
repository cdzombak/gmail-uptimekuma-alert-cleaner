# gmail-uptimekuma-alert-cleaner

Automatically clean up resolved Uptime Kuma alert pairs (down/up) from your Gmail inbox.

## What It Does

This program finds pairs of Uptime Kuma down/up alert emails in your Gmail inbox and cleans them up by:
- Removing the star (if present)
- Marking the messages as read
- Archiving them

Only messages matching all of these criteria are processed:
- In your Inbox
- Labeled `DzOps/Alerts`
- From sender "Uptime Kuma"
- Within the configured time window (default: 24 hours)
- Form complete down/up pairs for the same service

## How It Works

The program works backwards in time to find matching down/up pairs:

- **Simple pair**: `Down → Up` - both messages are cleaned up
- **Multiple downs**: `Down → Down → Up` - only the most recent down and the up are paired
- **Recent down remains**: `Down → Up → Down` - only the older pair is cleaned, the most recent down stays in your inbox
- **Multiple services**: Each service is tracked independently

## Installation

### Build from source

```bash
go install github.com/cdzombak/gmail-uptimekuma-alert-cleaner@latest
```

Or clone and build:

```bash
git clone https://github.com/cdzombak/gmail-uptimekuma-alert-cleaner.git
cd gmail-uptimekuma-alert-cleaner
go build
```

## Setup

### 1. Gmail API Credentials

Create a Google Cloud Platform project with Gmail API access:

1. Go to the [Gmail API Go Quickstart](https://developers.google.com/gmail/api/quickstart/go)
2. Follow "Step 1: Enable the Gmail API"
3. Download the `credentials.json` file

### 2. Configuration Directory

Create a configuration directory and place your credentials file there:

```bash
mkdir -p ~/.config/gmail-uptimekuma-alert-cleaner
mv ~/Downloads/credentials.json ~/.config/gmail-uptimekuma-alert-cleaner/
```

### 3. Optional: Create Config File

Copy the example config file to customize settings:

```bash
cp config.example.json ~/.config/gmail-uptimekuma-alert-cleaner/config.json
```

Edit `config.json` to customize the time window:

```json
{
  "timeWindow": "48h"
}
```

Valid time units: `h` (hours), `m` (minutes), `s` (seconds). Default is `24h` if no config file is present.

### 4. Initial Authorization

Run the program interactively the first time to authorize Gmail access:

```bash
gmail-uptimekuma-alert-cleaner \
  -configDir ~/.config/gmail-uptimekuma-alert-cleaner \
  -dry-run true
```

This will:
1. Open a browser for OAuth authorization
2. Save the token for future use
3. Show you what would be cleaned up (dry-run mode)

## Usage

### Command-line Flags

- `-configDir <path>` - Path to configuration directory (required if `GMAIL_UPTIMEKUMA_ALERT_CLEANER_CONFIG_DIR` env var is not set)
- `-dry-run <bool>` - Dry-run mode (default: `true`). Set to `false` to actually modify emails
- `-verbose` - Enable verbose debug logging
- `-version` - Print version and exit

### Dry-run Mode (Default)

By default, the program runs in dry-run mode and only prints what it would do:

```bash
gmail-uptimekuma-alert-cleaner -configDir ~/.config/gmail-uptimekuma-alert-cleaner
```

Example output:
```
DRY RUN MODE - would clean up the following message pairs:

Pair 1 - api.example.com:
  Down: 2025-01-24T10:15:30Z (ID: abc123)
  Up:   2025-01-24T10:20:45Z (ID: def456)
  Actions: unstar (if starred), mark read, archive

Total: 1 pairs (2 messages)
```

### Actually Clean Up Messages

To actually modify your emails, disable dry-run mode:

```bash
gmail-uptimekuma-alert-cleaner \
  -configDir ~/.config/gmail-uptimekuma-alert-cleaner \
  -dry-run false
```

### Running with Cron

Example crontab entry to run daily at 2 AM:

```cron
GMAIL_UPTIMEKUMA_ALERT_CLEANER_CONFIG_DIR=/home/user/.config/gmail-uptimekuma-alert-cleaner
0 2 * * * /usr/local/bin/gmail-uptimekuma-alert-cleaner -dry-run false
```

## How Pairs Are Matched

The program processes messages chronologically (oldest to newest) and pairs them according to these rules:

1. Each message can only be part of one pair
2. A Down message is tracked for each service
3. When an Up message is found, it's paired with the most recent unpaired Down for that service
4. Both messages in the pair are marked for cleanup

### Examples

**Example 1: Simple resolved issue**
```
Timeline: Down(ServiceA) → Up(ServiceA)
Result:   Clean up both messages
```

**Example 2: Multiple downs before recovery**
```
Timeline: Down1(ServiceA) → Down2(ServiceA) → Up(ServiceA)
Result:   Clean up Down2 and Up; Down1 remains unpaired
```

**Example 3: Service is currently down**
```
Timeline: Down1(ServiceA) → Up(ServiceA) → Down2(ServiceA)
Result:   Clean up Down1 and Up; Down2 stays in inbox (service is down)
```

**Example 4: Multiple resolved incidents**
```
Timeline: Down1(ServiceA) → Up1(ServiceA) → Down2(ServiceA) → Up2(ServiceA)
Result:   Clean up both pairs: (Down1, Up1) and (Down2, Up2)
```

**Example 5: Interleaved services**
```
Timeline: Down(ServiceA) → Down(ServiceB) → Up(ServiceA) → Up(ServiceB)
Result:   Clean up both pairs: (ServiceA: Down, Up) and (ServiceB: Down, Up)
```

## Logging

The program uses structured logging to stderr:

- **Normal mode**: Info-level messages about what's being processed
- **Verbose mode** (`-verbose`): Debug-level messages for troubleshooting
- **Dry-run output**: Human-readable output goes to stdout

## Exit Codes

The program uses semantic exit codes from [exitcode_go](https://github.com/cdzombak/exitcode_go):

- `0` - Success
- `2` - Invalid arguments
- `69` - Service unavailable (Gmail API failure)
- `74` - I/O error
- `78` - Configuration error

## License

MIT License - see LICENSE file for details.

## Author

Chris Dzombak
- [GitHub](https://github.com/cdzombak)
- [Website](https://www.dzombak.com)

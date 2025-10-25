# gmail-uptimekuma-alert-cleaner specification

## Background

Uptime Kuma is an open-source uptime monitoring program that can produce email alerts when a service goes down or comes back up. These emails come from a sender called "Uptime Kuma" — sender address varies — and their subjects start with either `✅ Up:` or `🔴 Down:`. The remainder of the subject is the service name.

In my particular setup, these emails appear in my Inbox and are always labeled `DzOps/Alerts`.

## Problem Statement

When something goes down and comes back up, I am left with pairs of down/up messages in my inbox, which I manually have to mark read, unstar, and archive, like the following:

![image-20251025115607217](/Users/cdzombak/code/gmail-uptimekuma-alert-cleaner/image-20251025115607217.png)

## Core Program Functionality

`gmail-uptimekuma-alert-cleaner` is a Go program that finds these pairs of messages, removes the star (if there is one), marks the message read (if necessary), and archives them. The end result is that these pairs of down/up messages end up marked read and archived in the `DzOps/Alerts` label.

The program *only* operates on messages matching the following criteria:

- in my Inbox
- from a sender whose display name is "Uptime Kuma"
- the label `DzOps/Alerts` is applied
- matching down/up pairs for the same service, working backwards in time

The last point is key. Note that time of arrival is critical:

- given the sequence "2 Down emails then 1 Up email" for a given service, *only* the Up email and the most recent Down email are cleaned up.
- given the sequence "1 Down email, 1 Up email, 1 Down email" for a given service, *only* the older Down/Up email pair is cleaned up. The most recent Down email remains in my inbox.

Note that the service covered by the email is also critical. Given the sequence "Service Y down, Service X down, Service Y up," *only* the Service Y emails are cleaned up.

## Additional Requirements

### Test Coverage

Core program logic (e.g. processing email sequences and deciding what to clean up) should be separated from Gmail interaction code and unit-tested.

### Dry-run Mode

Dry-run mode is the default. In this mode, `gmail-uptimekuma-alert-cleaner` accesses my Gmail account but only *prints* what it would change, in human-readable form, to standard output; it does not actually touch any emails.

The user disables dry-run mode by passing `-dry-run false` to the program.

### Reference Implementation

My program https://github.com/cdzombak/gmail-cleaner already performs some Gmail cleanup actions. Refer to how it is configured and run, and mirror that -- I should be able to use the same OAuth credentials and config file format.

### One-shot

This program will be run by cron; therefore it does not need any sort of daemon-style persistence.

### Logging

This program provides reasonable, helpful logging by default, and passing `-verbose` allows it to print more in-depth logging for debug purposes. The program uses the Go standard library's `slog` package to print logs to standard error.

### Time Window

The program operates on emails from the last 24 hours; this is configurable via the config file.

### Edge cases

#### Multiple Up emails following a single Down for the same service?

Only the first Up email following a single Down email for the same service is cleaned up. This program operates *only* on pairs of messages; it never cleans up multiple Up emails for a single Down email, or vice versa.

#### Messages at the boundary of our search window that might have pairs outside it?

*Only* examine and clean up messages within the search time window.

### Exit codes

Agent/Claude: Please use https://github.com/cdzombak/exitcode_go and choose semantically-relevant error codes for various errors, but don't go overboard.
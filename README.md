# whook

A self-hostable webhook gateway. It sits in front of your application, captures
every inbound webhook the moment it arrives, returns a fast 2xx so the provider
is happy, and then forwards the event to one or more destinations with automatic
retries. You can list, inspect, and replay any captured event.

## The problem

Services like Stripe, GitHub, and Shopify notify you about things that happen by
sending an HTTP POST to a URL you give them. A payment succeeds, a pull request
opens, an order is placed, and they POST the details to your server. That POST is
the webhook.

The catch is that the POST arrives exactly once, at a moment you do not control,
and pointing a provider straight at your application is fragile:

- If your app is down, deploying, or briefly crashing when the webhook lands, the
  event is lost. Many providers retry weakly or not at all, so a "payment
  succeeded" or "new order" event silently disappears and your data drifts from
  reality.
- When processing fails, the original request is already gone. There is no record
  of what the provider actually sent, so debugging is blind.
- A provider usually allows only one destination URL, but several internal
  services may need the same event (billing, email, analytics). That forces
  hand-written glue to receive it once and fan it out.
- There is no simple way to replay a past webhook to reproduce a bug or recover a
  missed event.

## Why a gateway helps

Instead of pointing the provider at your app, you point it at the gateway. The
gateway owns the inbound webhook end to end:

1. It saves every request to disk the instant it arrives, then immediately tells
   the provider "got it." The event is safe even if your app is down.
2. It forwards the saved event to your app on its own schedule. If your app is
   unreachable, it waits and retries with growing backoff until it succeeds or
   gives up and sets the event aside in a dead-letter pile.
3. It can forward one event to many destinations, each tracked on its own.
4. It keeps a durable copy of every request so you can list, inspect, and replay
   anything that came in.

Hosted products solve this, but they cost money and you hand them your traffic.
The open-source options mostly cover the opposite direction (sending webhooks to
your own customers). whook is for the receiving side, and you run it yourself.

## Stack

- Go, single static binary
- SQLite for storage (pure-Go driver, no cgo)
- Standard library HTTP, server-rendered UI
- dbmate for migrations

## Development

```sh
go build ./...
go test ./...
```

Run the gateway (defaults to a local SQLite file):

```sh
go run ./cmd/whook
```

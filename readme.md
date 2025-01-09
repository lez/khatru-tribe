# Khatru-tribe

A relay that only accepts events signed by tribe members.

A tribe is a hierarchical community that is defined by a leader and a tribe id.

NIP: https://github.com/lez/nipls/blob/main/tribe.md

# Install

```
go install github.com/lez/khatru-tribe@latest
```
Make sure the binary `khatru-tribe` is in your $PATH. It should be somewhere in `~/go/bin`.

# Run

```
khatru-tribe <leader-pubkey> <tribeid>
```
If successful, it starts listening on port 3334.

To add somebody to the tribe:
```
nak event -k 77 -t c=<tribeid> -t p=<new-member-pubkey> --sec <leader-nsec> ws://127.0.0.1:3334
```
The new member can now post to the relay and add more members to the tribe.

# Contributing

Contributions are welcome. At this point testing the stuff and providing feedback is a huge contribution!

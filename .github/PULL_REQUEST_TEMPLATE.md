## Description

<!-- What and why. One logical change per PR. -->

## Parity check

wgo must stay byte-for-byte compatible with WireViz 0.4.1. Confirm:

- [ ] `go test ./...` passes (including golden outputs)
- [ ] No output format, ordering, or console behavior changed
- [ ] If output logic changed, `make parity` or golden regen was run and the
      diff is explained in the PR description

## Checklist

- [ ] `make fmt` clean
- [ ] `make vet` clean
- [ ] Tests added/updated for the change

## Notes

<!-- Anything a reviewer should know: scope, trade-offs, follow-ups. -->

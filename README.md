# ext-gameboard

Public **contract surface** for the `gameboard` extension — the frozen, cross-repo boundary of
the **GameBoard.live** basketball game-day vertical. It follows the
[`extension-contract-repo`](https://github.com/sneat-co/sneat-libs/blob/main/spec/features/extension-contract-repo/README.md)
convention (mirrors [`contactus-ext`](https://github.com/sneat-co/contactus-ext)).

It holds only what the apps, the private `gameboard` impl repo, and other extensions need in order
to *talk to* gameboard — the TypeSpec wire contract, Go facade interfaces / DTOs / model shapes, and
the TS contract tokens/types. It contains **no** gameboard implementation (that lives in the private
[`sneat-co/gameboard`](https://github.com/sneat-co/gameboard) repo).

## Layout

```
ext-gameboard/
├── typespec/   # api4gameboard.tsp — the frozen wire contract (source of truth)
├── backend/    # Go module github.com/sneat-co/ext-gameboard/backend (hand-implemented to match the .tsp)
└── frontend/   # nx lib published as @sneat/extension-gameboard-contract (hand-implemented to match the .tsp)
```

## TypeSpec is the source of truth (no emitters)

Per the established house convention (`eventus/typespec`, `sneat-go/typespec`), the `.tsp` files are
the **frozen wire contract** and **no code emitters are configured**. The Go (`backend/`) and TS
(`frontend/`) sides **hand-implement matching types** against the `.tsp`. Shape/parity tests keep the
two language bindings in agreement with the contract. *(This realizes the master plan's "wire
TypeSpec Go/TS gen" item using the codebase's actual no-emitter pattern — see the gameboard-live
master plan.)*

## The load-bearing invariant

`ext-gameboard` depends **only on foundational/core code — never on another extension.** Because it
has no edge back to any sibling, `sibling → ext-gameboard` can never form a dependency cycle. An
interface or type belongs here **only if its entire signature is expressible in gameboard-own +
foundational/core types**. The CI check in `.github/workflows/ci.yml`
(`scripts/check-no-extension-deps.sh`) enforces the invariant.

## Source spec

Backstage feature tree [`sports/gameboard-live`](https://github.com/sneat-co/backstage/tree/main/spec/features/sports/gameboard-live)
and the [`gameboard-live` master plan](https://github.com/sneat-co/backstage/blob/main/spec/plans/gameboard-live.md).
The foundation contract here is the **event-timeline** record (append-only, idempotent by
client-generated `eventID`, ordered by server wall-clock with ties broken by `eventID`, deterministic
fold).

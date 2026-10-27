# Weekly evidence record

Student number:100854079
Week:3
Repository URL:
Immutable tag:e2935dbb505d76951bd9f9a7df88553f2efa8bbc
Commit hash:e2935dbb505d76951bd9f9a7df88553f2efa8bbc
Go version / OS:

## Commands actually executed
go run ./cmd/request -url http://127.0.0.1:8081/v1/peers
go run ./cmd/request -url http://127.0.0.1:8081/v1/peers -method POST -file testdata/peer-b.json -expect 201
go run ./cmd/request -url http://127.0.0.1:8081/v1/peers -method POST -file testdata/peer-b.json -expect 409
go run ./cmd/request -url http://127.0.0.1:8081/v1/peers
go run ./cmd/request -url http://127.0.0.1:8081/v1/files
go run ./cmd/request -url http://127.0.0.1:8082/v1/files
go run ./cmd/request -url http://127.0.0.1:8082/v1/peers
go run ./cmd/request -url http://127.0.0.1:8081/healthz
go run ./cmd/request -url http://127.0.0.1:8081/v1/status
go test ./internal/state -run "^TestWeek03_" -v -count=1
go test -race ./... -count=1 -timeout 30s
## Observed results (do not paste expected examples as results)
[] (empty registry on A)
201 with the peer B record
409 {"error":"duplicate_peer"}
A lists exactly one peer, node-b
A’s files: only alpha.txt
B’s files: diagram.txt then notes.txt. The fixture lists notes first, so seeing diagram first proves your catalogue sorts.
B’s peers are still [], which proves posting to A copied nothing into B.
and 9. /healthz and /v1/status still work, so Weeks 1 and 2 still pass.
## Failure case and diagnosis

## Design decision and alternative

## Remaining limitations

## Sources and AI-use disclosure

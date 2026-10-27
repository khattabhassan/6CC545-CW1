# Week 3: Architecture and Design Rationale

## 1. Architecture and Ownership

The application is organised around three main state objects: `Status`, `Registry`, and `Catalog`. Each node creates and manages its own instances through the application composition functions. Node A and Node B therefore have separate in-memory state, even when they run on the same computer.

The `Status` object stores the node's current operational status. The `Registry` stores peer records, while the `Catalog` stores metadata about files. Each object is responsible for protecting its own mutable state with a read-write mutex. HTTP handlers interact with these objects through their public methods rather than directly accessing internal maps or slices.

The registry's `Add` method controls peer insertion, while `Snapshot` provides a copy of the current peer records. Similarly, the catalogue uses `Replace` to validate and update its collection and `Snapshot` to return an independent copy. This separation keeps state management, validation, and HTTP request handling distinct.

At startup, `loadSeed` reads the configured seed file and passes its metadata to `Catalog.Replace`. This allows each node to initialise its catalogue from its own configuration without sharing mutable memory with another node.

## 2. Local State Is Not Global Truth

A peer record represents information stored by the local node, not necessarily the peer's actual current condition. For example, if Node A records Node B as `online`, this is a stored claim rather than proof that Node B is currently reachable.

Peer validation checks whether the record has acceptable syntax and values. It does not contact the peer or measure its availability. The current design does not provide a heartbeat mechanism, automatic failure detection, or gossip protocol to keep peer records synchronised.

Consequently, Node A and Node B can have different views of the network. A peer registered on Node A does not automatically appear in Node B's registry. This behaviour is consistent with the design because each process owns its own state.

## 3. Why the Duplicate Check Must Be Inside the Write Lock

The registry must prevent two concurrent requests from registering the same peer ID. The duplicate check and insertion therefore take place within a single critical section protected by the write lock.

If the check and insertion were separate, the following race could occur:

1. Writer A checks the registry and finds that the ID does not exist.
2. Writer B checks before Writer A inserts the record and also finds that the ID does not exist.
3. Writer A inserts the record.
4. Writer B inserts the same ID, potentially overwriting the first record or violating the uniqueness requirement.

With the duplicate check and insertion protected by the same write lock, only one writer can execute that section at a time. The first valid writer inserts the record; subsequent writers find the ID already present and receive `ErrDuplicate`.

The expected concurrency test launches 50 goroutines attempting to add the same ID. Exactly one should succeed, while the other 49 should report duplicates. This tests the uniqueness guarantee under simultaneous attempts.

## 4. Why Snapshots Copy and Sort

A snapshot should not expose the live data structures to callers. Returning an independent copy prevents a caller from modifying the registry's returned slice or the catalogue's returned slice and unintentionally changing the stored collection.

Sorting the copy provides deterministic output. A Go map does not guarantee iteration order, so sorting registry results by peer ID ensures that equivalent registry contents produce consistently ordered results. Sorting catalogue entries by file name has the same benefit.

These properties improve test reliability, make API responses easier to compare, and reduce unexpected behaviour for clients. An empty registry should also return an empty slice so that its JSON representation is `[]` rather than `null`.

The `Files` handler uses `Catalog.Snapshot` instead of accessing the catalogue's internal fields. The catalogue owns its mutex and storage, so it is responsible for providing a consistent copy. This encapsulation prevents HTTP handlers from bypassing the locking rules or depending on the internal representation.

## 5. What the Race Test Demonstrates

The race test checks whether multiple goroutines can attempt to insert the same peer ID without violating the uniqueness rule. A successful result demonstrates that, for the tested execution, one insertion succeeds and the remaining attempts are rejected as duplicates.

The Go race detector can identify certain unsynchronised memory accesses encountered while the instrumented tests execute. However, a successful run does not prove that the program is free of every possible concurrency defect. It only provides evidence about the executions and code paths that were actually tested.

The normal concurrent test and the race-detector run should therefore be recorded separately. If the race detector cannot run because the required compiler or CGO environment is unavailable, that limitation should be documented rather than claiming the race test passed.

## 6. File Metadata and SHA-256 Limitations

The file catalogue stores metadata, including the file name, size, modification time, and SHA-256 value. Validation can check that the hash contains exactly 64 hexadecimal characters and reject malformed values.

However, a syntactically valid hash does not prove that the corresponding file bytes were received, stored, or verified. The catalogue manages metadata rather than file contents, and a hash string supplied by a client could be incorrect or unrelated to any actual file.

To establish integrity, a system would need access to the relevant file bytes, calculate their SHA-256 digest, and compare the calculated digest with the claimed value. That verification is outside the current metadata-only catalogue's responsibility.

## 7. Summary

The Week 3 design separates application composition, HTTP handlers, data validation, and state management. Each node owns its own status, registry, and catalogue. Mutexes protect concurrent access, while snapshots provide independent and consistently ordered views of stored data.

The design provides a foundation for later networking and file-processing features, but it does not yet establish a globally consistent view of peers or verify the underlying contents of catalogued files. These limitations are important when interpreting API responses and describing what the implementation guarantees.
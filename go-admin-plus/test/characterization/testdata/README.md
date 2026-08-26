# Pre-phase-one SQLite fixture

`pre_phase1.sql` is the logical SQLite snapshot used by migration characterization.
It represents the public data shape immediately before the phase-one host refactor:

- migration version `1786700000000` (`1786700000000_demo_product.go`) has been applied;
- one live `demo_product` row exercises identity, ownership, timestamps, soft delete,
  text, decimal, and status preservation;
- all values are synthetic and deterministic.

The fixture was derived from the migration and model definitions at backend source
commit `750537c0b8edde522f9c9dddc2bfcf64169689ea`. It is SQL rather than a binary
database so its provenance and changes remain reviewable. `fixture_test.go` pins its
SHA-256 digest; later migration tests should import this script into a temporary
SQLite database and must not update the digest without reviewing the data contract.

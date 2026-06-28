# kaschnit/immut

Immutable Go data structure library focused on high performance.

## Data structures

### Map

Map is an immutable map based on the Compressed Hash-Array Mapped Prefix-tree (CHAMP) data structure: https://michael.steindorfer.name/publications/oopsla15.pdf.

The reference implementation of CHAMP can be seen in the Capsule library for Java: https://github.com/usethesource/capsule.

## Development

### Getting Started

#### Prerequisites
- go version v1.21+ (go toolchains support)

### Run unit tests

Run `make test`.

### Run benchmark tests

Run `make test-bench`.

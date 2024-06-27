# metal-cli

To install dependencies:

```bash
bun install
```

To run:

```bash
bun run index.ts
```

This project was created using `bun init` in bun v1.1.1. [Bun](https://bun.sh) is a fast all-in-one JavaScript runtime.

## Running locally in another directory

- First, in this repository's directory, run `bun link`. This registers the directory as a linkable package. This allows you to bring the local version of a CLI to a project's `node_modules/.bin` with the command:

- `bun link metal-cli`. Go ahead and do this in this directory you want to run `metal` in so that you have `metal` accessible as a command to run.

- Make sure `./node_modules/.bin` is in your PATH so that you can just type `metal`

## Developing locally

`bun run dev <command>`, e.g. `bun run dev loginz`.

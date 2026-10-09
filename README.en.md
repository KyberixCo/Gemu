# Gemu · by Kyberix

**Your PC, your library, your console experience.**

[Español](README.md) · English

Gemu aims to turn a Windows PC into a living-room gaming console: power on the computer, choose a game with a controller, and start playing through a single interface designed for televisions.

This document describes the product vision and its goals.

## Play from the couch

The experience should be comfortable at a distance: readable text, large artwork, clear focus, and predictable navigation with a D-pad or stick. Each screen should focus on one task, with accessible actions and minimal scrolling.

Transitions, selection glow, and sound should help users find their way and confirm actions. Users should be able to adjust or disable motion and audio while keeping navigation clear.

## One unified library

Gemu seeks to bring games from Steam, Epic Games, GOG, standalone executables, and emulation libraries into one catalog. Connectors should be interchangeable and let users choose which sources to use.

A game should have its own identity, even when it exists across multiple stores or installations. Favorites, search, and collections should help users find something to play without browsing each client separately. The library and locally stored data should remain available offline.

Launching should respect official clients and their authentication and DRM mechanisms. Emulators, ROMs, and BIOS files are configured as external resources supplied by the user.

## Launch a game and return to Gemu

The complete gaming session is a priority: prepare the launch, hand focus to the game, and restore the interface when it ends, preserving the previous selection.

Gemu should stop interpreting controller navigation during gameplay. Launch failures and disconnected devices should offer a clear way forward. Session history and duration should rely on verifiable evidence that a game is running.

## Management designed for television

Library, emulator, and controller settings should follow the same console interaction model: small groups of settings, editing one value at a time, and readily accessible Back and Save actions.

The goals include controller profiles for each destination, button mapping, and calibration. Text entry should offer television-friendly options and support a keyboard when it is more convenient.

## Metadata, artwork, and assistance

Game details should be enriched through configurable metadata and artwork providers, with priorities, provenance, and credential management when an API requires it.

The assistant should help organize the library, prepare configurations, and propose improvements. Each proposal should show what changes and let users accept individual fields or discard it before saving. Users retain control over their data and changes.

## Technical foundation and reliability

The target architecture combines a Wails desktop application, Go services, a React and TypeScript interface, and local SQLite storage. The domain, connectors, and Windows integration should remain separate to support testing and additional providers.

Priorities include fast navigation, moderate resource usage, focus recovery, and data preservation. Automatic startup and a possible shell mode should be optional, reversible, and provide a recovery path.

## Documentation

The detailed documentation below is in Spanish.

- [Vision and architecture](idea.md)
- [Experience design](docs/design.md)
- [Motion and sound](docs/feedback.md)
- [Management and integration contracts](docs/management.md)
- [Development guide](docs/development.md)

## License

Gemu is licensed under the **GNU General Public License, version 3 only** (`GPL-3.0-only`). See [LICENSE](LICENSE). Third-party dependencies and assets retain their respective licenses.

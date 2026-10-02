# Sniper

Sniper is a compact Go desktop utility for Windows. Save reusable text, find it quickly, and copy it to the clipboard.

**Status:** fully specified on October 1, 2026, and ready for autonomous implementation. No code exists yet. Agents should start with [CLAUDE.md](CLAUDE.md).

## What it does

Open a small window, type to search your snippets, press Enter, and paste into your browser or any other application. You can create, edit, and delete snippets in the same window. Data is kept in a local JSON file (`%AppData%\Sniper\snippets.json`) with an automatic last-good backup, and you can import and export it.

The stack is Go with Fyne 2.8. It needs no browser extension, account, or network access.

## Documents

| Doc | Purpose |
|---|---|
| [CLAUDE.md](CLAUDE.md) | Operating instructions for the implementing agent |
| [01 Research](docs/01-research.md) | Rationale and verified toolkit facts |
| [02 Product spec](docs/02-product-spec.md) | Behavior, rules, limits, file format, acceptance criteria |
| [03 Architecture](docs/03-architecture.md) | Packages, APIs, persistence algorithm, startup |
| [04 Delivery plan](docs/04-delivery-plan.md) | Ordered milestones with exit checks |
| [05 UI spec](docs/05-ui-spec.md) | Layout, keyboard map, dialogs, exact strings |
| [06 Testing](docs/06-testing.md) | Automated tests, the clipboard check, human checks |
| [07 Decision log](docs/07-decision-log.md) | Decisions and measurements |

Direct insertion, sync, templates, tray, and global hotkeys are outside the MVP.

# alttpr-go

A Go CLI port of the [ALttP VT Randomizer](https://github.com/sporchia/alttp_vt_randomizer) (Laravel/PHP). Single static binary, no PHP runtime, no database, no web server.

## Status

Full randomizer pipeline ported and validated:

- ROM open / base-patch apply / checksum / save
- World / Region / Location / Item / Boss models
- Filler (RandomAssumed)
- Prize placement with `prize.crossWorld` (Shuffle Prizes), per-type shuffle flags
- Standard and Open game states
- Goals: `ganon`, `fast_ganon`, `dungeons`, `ganonhunt`, `triforce-hunt`, `pedestal`, `completionist`
- ROM setters: dialog, credits, initial SRAM, hints, item placement, mode flags
- Spoiler JSON output, playthrough analysis

Validated end-to-end in Mesen2 against a generated ROM — Link spawns at the
correct coordinates, the randomized intro renders, and the WRAM mirror
matches the InitialSram template byte-for-byte. See `internal/world/parity_test.go`
and `internal/randomizer/end_to_end_test.go` for the parity checks.

Not ported: ZSPR sprite injection, enemizer integration, tournament-mode
seed naming, the web UI / REST API (out of scope for the CLI port).

## Build

```sh
go build ./cmd/alttpr
```

Requires Go 1.26+. The base-patch JSON is embedded at compile time
(`internal/patch/all_patches_embed/edc01f3db798ae4dfe21101311598d44.json`),
so no runtime dependencies.

## Usage

```sh
# Unrandomized: base patch + QoL only, byte-identical to PHP --unrandomized
./alttpr --unrandomized base.sfc out/

# Default Standard / Ganon goal, with spoiler
./alttpr base.sfc out/

# Triforce hunt
./alttpr --goal=triforce-hunt --triforce-pieces=30 --triforce-goal=20 base.sfc out/

# Disable Shuffle Prizes (pendants stay in light-world dungeons)
./alttpr --shuffle-prizes=false base.sfc out/
```

`./alttpr --help` lists the full flag set.

`base.sfc` is the unmodified Japan 1.0 LoROM (`MD5 03a63945398191337e896e5771f77173`).
Bring your own — it is not redistributed here.

## Repo layout

```
cmd/alttpr/             CLI entrypoint
cmd/alttpr-brick/       Trimui Brick handheld harness (experimental)
internal/rom/           Binary ROM file: open, patch, checksum, save
internal/patch/         Embedded base-patch JSON + apply logic
internal/item/          Item model, registry, ItemCollection
internal/boss/          Boss registry
internal/logic/         Decoupling interface between item and world packages
internal/helpers/       Ports of app/Helpers/*.php (fy_shuffle, hash_array, etc.)
internal/world/         World, Region, Location, Spoiler, Playthrough,
                        Dialog, Credits, InitialSram, all ROM setters
internal/filler/        RandomAssumed filler
internal/randomizer/    Orchestrator (Randomize → placements → ROM write)
internal/job/           job.Options shared by CLI and embed callers
vendor/php-randomizer/  Submodule: the upstream PHP randomizer, used as a
                        reference for parity tests and to regenerate the
                        embedded base-patch JSON when its source changes.
```

## Working with the PHP reference

The submodule is the canonical PHP randomizer. Two reasons it's here:

**1. Regenerate the embedded base patch.** When the upstream randomizer's
ASM sources change (`vendor/z3/randomizer/*.asm`), the JSON patch must be
rebuilt:

```sh
cd vendor/php-randomizer
composer install
cp .env.example .env && php artisan key:generate
# Set ENEMIZER_BASE=/abs/path/to/base.sfc in .env
php artisan migrate
php artisan alttp:updatebuildrecord
cp storage/patches/*.json ../../internal/patch/all_patches_embed/
```

Then bump the patch filename referenced by `internal/patch/embed.go`.

**2. Regenerate parity fixtures.** Several Go tests compare against PHP
output:

- `internal/world/parity_test.go` — dialog encoder, InitialSram
- `internal/world/credits_test.go` — credits encoder
- `internal/world/parity_rom_regions_test.go` — per-region ROM byte ranges

These tests `t.Skip()` when fixtures are absent. To regenerate them you
run small PHP scripts inside the submodule that emit binaries to `/tmp/`
(scripts not yet checked in — write them ad-hoc or as needed).

## Testing

```sh
go test ./...
```

`internal/randomizer/end_to_end_test.go` runs the full pipeline (no PHP
dependency) and asserts every filled location's first byte matches the
placed item.

## License

The Go code in this repository is offered under the same license as the
upstream PHP randomizer (see `vendor/php-randomizer/LICENSE`).

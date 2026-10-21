# Packaging

Canonical copies of the external distribution manifests. The live copies live in
their own repositories:

- Homebrew: `SrIruma/homebrew-tap` (`Formula/repoctx.rb`)
- Scoop: `SrIruma/scoop-bucket` (`bucket/repoctx.json`)
- winget: submitted as a PR to `microsoft/winget-pkgs`

## Bumping on a new release

1. Download the new `SHA256SUMS.txt` from the release.
2. Update `version` and every `url`/`sha256` in:
   - `packaging/homebrew/repoctx.rb`
   - `packaging/winget/SrIruma.repoctx*.yaml`
   - `packaging/scoop/repoctx.json`
3. Copy the updated files into the tap/bucket repos and push.
4. For winget, open a new PR against `microsoft/winget-pkgs` with the updated
   manifest (or use `wingetcreate update SrIruma.repoctx --version X.Y.Z`).
5. npm is published automatically by the `publish-npm` job in
   `.github/workflows/release.yml` when `NPM_TOKEN` is set.

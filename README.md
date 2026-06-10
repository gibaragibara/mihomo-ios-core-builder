# Mihomo iOS Core Builder

This repository builds a minimal gomobile wrapper around `github.com/metacubex/mihomo` for use as an iOS linkable core.

It is intentionally separate from the Anywhere app repository. The public repository contains only the build wrapper and GitHub Actions workflow, not the app code.

## Output

The workflow uploads:

- `MihomoCore.xcframework.zip`

The framework is produced by:

```sh
gomobile bind -target=ios,iossimulator -prefix=MihomoCore -o build/MihomoCore.xcframework ./mihomocore
```

## Wrapper API

The Go package exports gomobile-compatible functions:

- `Start(configPath, workDirectory string) int32`
- `Stop()`
- `Reload(configPath string) int32`
- `Health() int32`
- `LastError() string`

Return code `0` means success. Non-zero means failure; call `LastError()` for the latest error message.

## Config Contract

The wrapper rejects configs that would conflict with Anywhere's first-stage integration:

- `mixed-port` must be `7890`
- `allow-lan` must not be `true`
- `tun.enable` must not be `true`
- `dns.enable` must not be `true`

This keeps mihomo as a local outbound SOCKS/mixed proxy only. Anywhere still owns TUN, TCP flow handling, DNS, and MITM.

## Build

Run the GitHub Actions workflow manually:

1. Open the repository on GitHub.
2. Go to `Actions`.
3. Select `Build Mihomo iOS Core`.
4. Run the workflow.
5. Download `MihomoCore.xcframework.zip` from the workflow artifacts.

## Notes

This uses gomobile's generated Objective-C/Swift bindings. The consuming Swift code should import the generated framework and call the generated API rather than relying on temporary raw C symbols.


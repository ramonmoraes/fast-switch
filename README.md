# Fast Switch

Truly go back to the previous application.

## macOS permissions

To get useful results, expect to grant:

- Accessibility, for focusing specific windows

## Releases

GitHub users should download the macOS release artifact `fast-switch-macos.zip` from the [repository's Releases page.](https://github.com/ramonmoraes/fast-switch/releases)

After downloading:

- unzip `fast-switch-macos.zip`
- move `fast-switch.app` to `/Applications`
- open the app from Applications

To create that release artifact locally:

```sh
make package
```

This writes the downloadable archive to `build/bin/fast-switch-macos.zip`.

## Development

```sh
wails dev
```

## Building
```sh
  wails build -clean
```

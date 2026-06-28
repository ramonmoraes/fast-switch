# Fast Switch

Truly go back to the previous application.


<a href="https://www.buymeacoffee.com/ramonmoraes" target="_blank"><img src="https://cdn.buymeacoffee.com/buttons/v2/default-yellow.png" alt="Buy Me a Coffee" style="height: 60px !important;width: 217px !important;" ></a>

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

# Fast Switch

Fast Switch helps you jump between <b>actual</b> projects and workflows instantly, so you
  can stay focused and keep momentum.

If you like the project, consider to 
<a href="https://www.buymeacoffee.com/ramonmoraes" target="_blank"><img src="https://cdn.buymeacoffee.com/buttons/v2/default-yellow.png" alt="Buy Me a Coffee" style="height: 24px !important;width: 96px !important;margin-left:16px;" ></a>


## Installing

GitHub users should download the macOS release artifact `fast-switch-macos.zip` from the [repository's Releases page.](https://github.com/ramonmoraes/fast-switch/releases)

After downloading:

- unzip `fast-switch-macos.zip`
- move `fast-switch.app` to `/Applications`
- open the app from Applications


### macOS permissions

To get useful results, expect to grant:

- Accessibility, for focusing specific windows

To create that release artifact locally:

```sh
make package
```

This writes the downloadable archive to `build/bin/fast-switch-macos.zip`.

## Development

```sh
wails dev # For live dev
wails build -clean # For building the binary
```

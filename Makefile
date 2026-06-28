APP_NAME := fast-switch
APP_BUNDLE := build/bin/$(APP_NAME).app
RELEASE_ZIP := build/bin/$(APP_NAME)-macos.zip
VERSION := $(shell git rev-parse --short HEAD)

.PHONY: build package release-gh

build:
	wails build -clean

package: build
	ditto -c -k --sequesterRsrc --keepParent "$(APP_BUNDLE)" "$(RELEASE_ZIP)"

release-gh: package
	gh release create "$(VERSION)" "$(RELEASE_ZIP)" --title "$(VERSION)" --notes ""

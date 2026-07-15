BUILD_DIR = build
BINARY    = yuanshu-ai
APP       = 元数AI.app
DMG       = 元数AI.dmg
PYTHON   ?= $(shell pyenv which python3 2>/dev/null || command -v python3)

# 当前平台的可执行文件名
ifeq ($(shell go env GOOS),windows)
BIN := $(BINARY).exe
else
BIN := $(BINARY)
endif

.PHONY: all mac win app dmg run clean build

all: mac win

build: clean all
	@$(MAKE) app 2>/dev/null || true
	@$(MAKE) dmg 2>/dev/null || true

mac:
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build \
		-ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY) .
	@echo "✅  $(BUILD_DIR)/$(BINARY)"

win: yuanshu-ai.syso
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build \
		-ldflags="-s -w -H windowsgui" -o $(BUILD_DIR)/$(BINARY).exe .
	@echo "✅  $(BUILD_DIR)/$(BINARY).exe"

yuanshu-ai.syso: yuanshu-ai.ico
	@which rsrc 2>/dev/null || go install github.com/akavel/rsrc@latest 2>&1
	rsrc -ico yuanshu-ai.ico -o yuanshu-ai.syso 2>/dev/null || true

run:
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BIN) .
	@echo "🚀 启动 $(BINARY)..."
	@$(BUILD_DIR)/$(BIN)

app: mac
	@rm -rf "$(BUILD_DIR)/$(APP)"
	@mkdir -p "$(BUILD_DIR)/$(APP)/Contents/MacOS" "$(BUILD_DIR)/$(APP)/Contents/Resources"
	@cp yuanshu-ai.icns "$(BUILD_DIR)/$(APP)/Contents/Resources/"
	@cp "$(BUILD_DIR)/$(BINARY)" "$(BUILD_DIR)/$(APP)/Contents/MacOS/"
	@printf '%s\n' \
		'<?xml version="1.0" encoding="UTF-8"?>' \
		'<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">' \
		'<plist version="1.0"><dict>' \
		"<key>CFBundleExecutable</key><string>$(BINARY)</string>" \
		"<key>CFBundleIdentifier</key><string>com.yuanshu.yuanshu-ai</string>" \
		"<key>CFBundleName</key><string>元数AI</string>" \
		"<key>CFBundleDisplayName</key><string>元数AI</string>" \
		"<key>CFBundleVersion</key><string>1.0</string>" \
		"<key>CFBundleIconFile</key><string>yuanshu-ai.icns</string>" \
		'</dict></plist>' \
	> "$(BUILD_DIR)/$(APP)/Contents/Info.plist"
	@echo "✅  $(BUILD_DIR)/$(APP)"

dmg: app
	@T=$$(mktemp -d); cp -R "$(BUILD_DIR)/$(APP)" $$T/; \
	ln -s /Applications $$T/Applications; \
	hdiutil create -volname "元数AI" -srcfolder $$T \
		-ov -format UDZO -size 100m "$(BUILD_DIR)/$(DMG)" >/dev/null 2>&1; \
	rm -rf $$T
	@echo "✅  $(BUILD_DIR)/$(DMG)"

icons:
	$(PYTHON) scripts/make_macos_icon.py
	@echo "✅  yuanshu-ai.icns"

clean:
	@rm -rf $(BUILD_DIR) yuanshu-ai.syso
	@echo "✅  已清理"

BUILD_DIR = build
BINARY    = yuanshu-ai
APP       = 元数AI.app
DMG       = 元数AI.dmg
GOOS     ?= $(shell go env GOOS)
PYTHON   ?= $(shell pyenv which python3 2>/dev/null || command -v python3)

.PHONY: all mac win app dmg run clean build

# 当前平台的二进制
all:
	@mkdir -p $(BUILD_DIR)
ifeq ($(GOOS),windows)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build \
		-ldflags="-s -w -H windowsgui" -o $(BUILD_DIR)/$(BINARY).exe .
	@echo "✅  $(BUILD_DIR)/$(BINARY).exe"
else
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build \
		-ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY) .
	@echo "✅  $(BUILD_DIR)/$(BINARY)"
endif

build: clean all

# macOS
mac:
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build \
		-ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY) .
	@echo "✅  $(BUILD_DIR)/$(BINARY)"

# Windows
win:
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build \
		-ldflags="-s -w -H windowsgui" -o $(BUILD_DIR)/$(BINARY).exe .
	@echo "✅  $(BUILD_DIR)/$(BINARY).exe"

# 运行当前平台二进制
run:
	@$(MAKE) all
	@echo "🚀 启动 $(BINARY)..."
ifeq ($(GOOS),windows)
	@$(BUILD_DIR)/$(BINARY).exe
else
	@$(BUILD_DIR)/$(BINARY)
endif

# macOS 应用包（仅 macOS）
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

# DMG（仅 macOS）
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
	@rm -rf $(BUILD_DIR)
	@echo "✅  已清理"

GO := go
DIST_DIR := dist
ROM_DIR := roms
WASM_GZIP_LIMIT := 3000000

.DEFAULT_GOAL := build

.PHONY: build
build: clean
	mkdir -p $(DIST_DIR)/roms
	GOOS=js GOARCH=wasm $(GO) build -trimpath -ldflags="-s -w -buildid=" -o $(DIST_DIR)/main.wasm ./web
	@WASM_EXEC="$$($(GO) env GOROOT)/lib/wasm/wasm_exec.js"; \
	if [ ! -f "$$WASM_EXEC" ]; then WASM_EXEC="$$($(GO) env GOROOT)/misc/wasm/wasm_exec.js"; fi; \
	cp "$$WASM_EXEC" $(DIST_DIR)/wasm_exec.js
	cp web/index.html $(DIST_DIR)/index.html
	cp web/favicon.svg $(DIST_DIR)/favicon.svg
	$(GO) run ./tools/romcatalog "$(ROM_DIR)" "$(DIST_DIR)/roms"
	@RAW_BYTES="$$(wc -c < $(DIST_DIR)/main.wasm)"; \
	GZIP_BYTES="$$(gzip -9 -c $(DIST_DIR)/main.wasm | wc -c)"; \
	printf 'WASM: %s bytes raw, %s bytes gzip\n' "$$RAW_BYTES" "$$GZIP_BYTES"; \
	if [ "$$GZIP_BYTES" -gt "$(WASM_GZIP_LIMIT)" ]; then \
		echo "Error: compressed WASM exceeds $(WASM_GZIP_LIMIT) bytes"; \
		exit 1; \
	fi
	@if find $(DIST_DIR) -type f \( -name '*_test.go' -o -name '*.test' -o -name '*.test.*' -o -name '*.spec.*' \) -print -quit | grep -q .; then \
		echo "Error: test files must not be included in $(DIST_DIR)"; \
		exit 1; \
	fi
	@find $(DIST_DIR) -type f -printf '%s %p\n' | sort -n

.PHONY: web
web: build

.PHONY: serve
serve: build
	python3 -m http.server 8080 --directory $(DIST_DIR)

.PHONY: web-serve
web-serve: serve

.PHONY: romInfo
romInfo:
	@if [ -z "$(ROM_FILE)" ]; then \
		echo "Error: ROM_FILE parameter required"; \
		echo "Usage: make romInfo ROM_FILE=<path/to/rom.nes>"; \
		exit 1; \
	fi
	@if [ -f "$(ROM_FILE)" ]; then \
		echo "ROM File: $(ROM_FILE)"; \
		echo "File Size: $$(stat -f%z "$(ROM_FILE)" 2>/dev/null || stat -c%s "$(ROM_FILE)" 2>/dev/null) bytes"; \
		echo "NES Header:"; \
		xxd -l 16 -g 1 "$(ROM_FILE)" | head -1; \
		CONTROL1=$$(printf "%d" 0x$$(xxd -p -l 1 -s 6 "$(ROM_FILE)")); \
		CONTROL2=$$(printf "%d" 0x$$(xxd -p -l 1 -s 7 "$(ROM_FILE)")); \
		MAPPER=$$((CONTROL2 & 0xF0 | CONTROL1 >> 4)); \
		MIRROR=$$((CONTROL1 & 1)); \
		BATTERY=$$((CONTROL1 >> 1 & 1)); \
		TRAINER=$$((CONTROL1 >> 2 & 1)); \
		FOURSCREEN=$$((CONTROL1 >> 3 & 1)); \
		PRG_SIZE=$$(printf "%d" 0x$$(xxd -p -l 1 -s 4 "$(ROM_FILE)")); \
		CHR_SIZE=$$(printf "%d" 0x$$(xxd -p -l 1 -s 5 "$(ROM_FILE)")); \
		echo "Mapper: $$MAPPER"; \
		if [ $$MIRROR -eq 0 ]; then echo "Mirroring: Horizontal"; else echo "Mirroring: Vertical"; fi; \
		if [ $$BATTERY -eq 1 ]; then echo "Battery RAM: Yes"; else echo "Battery RAM: No"; fi; \
		if [ $$TRAINER -eq 1 ]; then echo "Trainer: Yes"; else echo "Trainer: No"; fi; \
		if [ $$FOURSCREEN -eq 1 ]; then echo "Four-screen: Yes"; else echo "Four-screen: No"; fi; \
		echo "PRG ROM: $$((PRG_SIZE * 16)) KB"; \
		echo "CHR ROM: $$((CHR_SIZE * 8)) KB"; \
	else \
		echo "Error: ROM file '$(ROM_FILE)' not found"; \
		exit 1; \
	fi

.PHONY: clean
clean:
	rm -rf $(DIST_DIR)
	rm -f web/main.wasm web/wasm_exec.js

.PHONY: help
help:
	@echo "GoNES browser emulator"
	@echo ""
	@echo "Available targets:"
	@echo "  make build                       - Build the minimal release in dist/"
	@echo "  make serve                       - Build and serve it on port 8080"
	@echo "  make web                         - Alias for make build"
	@echo "  make web-serve                   - Alias for make serve"
	@echo "  make romInfo ROM_FILE=<file>     - Display ROM header information"
	@echo "  make clean                       - Remove generated release files"
	@echo "  make help                        - Display this help message"

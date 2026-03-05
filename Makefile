WHISPER_CPP_DIR := third_party/whisper.cpp
WHISPER_LIB := $(WHISPER_CPP_DIR)/build/src/libwhisper.a
WHISPER_INCLUDE_DIR := $(WHISPER_CPP_DIR)/include
GGML_INCLUDE_DIR := $(WHISPER_CPP_DIR)/ggml/include
GGML_LIB_DIR := $(WHISPER_CPP_DIR)/build/ggml/src

MODEL_DIR := $(HOME)/.local/share/sussurai/models
DEFAULT_MODEL := $(MODEL_DIR)/ggml-base.bin

BIN_DIR := $(HOME)/.local/bin
APP_DIR := $(HOME)/.local/share/applications
ICON_DIR := $(HOME)/.local/share/icons/hicolor/128x128/apps
AUTOSTART_DIR := $(HOME)/.config/autostart

.PHONY: all clean download-model test install uninstall

all: sussurai

# Clone whisper.cpp if not present
$(WHISPER_CPP_DIR)/CMakeLists.txt:
	@echo ">>> Cloning whisper.cpp..."
	mkdir -p third_party
	git clone --depth 1 https://github.com/ggml-org/whisper.cpp.git $(WHISPER_CPP_DIR)

# Build whisper.cpp static library
$(WHISPER_LIB): $(WHISPER_CPP_DIR)/CMakeLists.txt
	@echo ">>> Building whisper.cpp..."
	cmake -S $(WHISPER_CPP_DIR) -B $(WHISPER_CPP_DIR)/build \
		-DCMAKE_BUILD_TYPE=Release \
		-DBUILD_SHARED_LIBS=OFF \
		-DWHISPER_BUILD_EXAMPLES=OFF \
		-DWHISPER_BUILD_TESTS=OFF
	cmake --build $(WHISPER_CPP_DIR)/build --config Release -j$$(nproc)

# Build sussurai
sussurai: $(WHISPER_LIB)
	@echo ">>> Building sussurai..."
	CGO_ENABLED=1 \
	C_INCLUDE_PATH=$(CURDIR)/$(WHISPER_INCLUDE_DIR):$(CURDIR)/$(GGML_INCLUDE_DIR) \
	LIBRARY_PATH=$(CURDIR)/$(WHISPER_CPP_DIR)/build/src:$(CURDIR)/$(GGML_LIB_DIR) \
	go build -o sussurai .

# Download default whisper model
download-model:
	@mkdir -p $(MODEL_DIR)
	@if [ ! -f "$(DEFAULT_MODEL)" ]; then \
		echo ">>> Downloading ggml-base.bin (~142MB)..."; \
		curl -L -o "$(DEFAULT_MODEL)" \
			"https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-base.bin"; \
		echo ">>> Model saved to $(DEFAULT_MODEL)"; \
	else \
		echo ">>> Model already exists at $(DEFAULT_MODEL)"; \
	fi

test: $(WHISPER_LIB)
	CGO_ENABLED=1 \
	C_INCLUDE_PATH=$(CURDIR)/$(WHISPER_INCLUDE_DIR):$(CURDIR)/$(GGML_INCLUDE_DIR) \
	LIBRARY_PATH=$(CURDIR)/$(WHISPER_CPP_DIR)/build/src:$(CURDIR)/$(GGML_LIB_DIR) \
	go test -v ./...

install: sussurai
	@echo ">>> Installing sussurai..."
	mkdir -p $(BIN_DIR) $(APP_DIR) $(ICON_DIR) $(AUTOSTART_DIR)
	cp sussurai $(BIN_DIR)/sussurai
	cp icons/sussurai.png $(ICON_DIR)/sussurai.png
	cp sussurai.desktop $(APP_DIR)/sussurai.desktop
	cp sussurai.desktop $(AUTOSTART_DIR)/sussurai.desktop
	@if [ -f .env ]; then \
		mkdir -p $(HOME)/.config/sussurai; \
		cp .env $(HOME)/.config/sussurai/.env; \
		echo ">>> Copied .env to ~/.config/sussurai/"; \
	fi
	@echo ">>> Installed! Sussurai is in your app menu and will auto-start on login."

uninstall:
	@echo ">>> Uninstalling sussurai..."
	rm -f $(BIN_DIR)/sussurai
	rm -f $(APP_DIR)/sussurai.desktop
	rm -f $(AUTOSTART_DIR)/sussurai.desktop
	rm -f $(ICON_DIR)/sussurai.png
	@echo ">>> Uninstalled."

clean:
	rm -f sussurai
	rm -rf $(WHISPER_CPP_DIR)/build

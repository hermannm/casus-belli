APP_NAME := bfh-server

PACKAGES := local public
PACKAGE_DIR := ./cmd
OUTPUT_DIR := bin

PLATFORMS := darwin linux windows
ARCHITECTURES := 386 amd64 arm64
EXCLUDED := darwin-386

# For every package, builds a binary for every given platform and architecture, except those in EXCLUDED.
# Adds .exe for Windows builds.
crosscompile:
	@$(foreach PACKAGE, $(PACKAGES),\
		$(foreach OS, $(PLATFORMS),\
			$(foreach ARCH, $(ARCHITECTURES),\
				$(if $(findstring $(OS)-$(ARCH), $(EXCLUDED)),,\
					echo "\033[0;35m[Building]\033[0m $(APP_NAME)_$(PACKAGE)_$(OS)-$(ARCH)$(if $(findstring windows, $(OS)),.exe,)";\
					env GOOS=$(OS) GOARCH=$(ARCH)\
					go build\
					-o $(OUTPUT_DIR)/$(PACKAGE)/$(APP_NAME)_$(PACKAGE)_$(OS)-$(ARCH)$(if $(findstring windows, $(OS)),.exe,)\
					$(PACKAGE_DIR)/$(PACKAGE);\
				)\
			)\
		)\
	)

	@echo "\033[0;32m[Finished]\033[0m Output in: $(OUTPUT_DIR)/"

# Avoids make clashing with potential file called crossocompile.
.PHONY: crosscompile

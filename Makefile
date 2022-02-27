APP := bfh-server

PLATFORMS := darwin linux windows
ARCHITECTURES := 386 amd64 arm64

PACKAGES := local public
PACKAGE_DIR := ./app
OUTPUT_DIR := bin

default: build
build:
# For every package, builds a binary for every given platform and architecture.
# Does not include 386 architecture for darwin.
# Adds .exe for Windows builds.
	@$(foreach PACKAGE, $(PACKAGES),\
		$(foreach OS, $(PLATFORMS),\
			$(foreach ARCH,\
				$(if $(findstring darwin, $(OS)),\
					$(filter-out 386, $(ARCHITECTURES)),\
					$(ARCHITECTURES)\
				),\
				echo "\033[0;36m[Building]\033[0m $(APP)_$(PACKAGE)_$(OS)-$(ARCH)$(if $(findstring windows, $(OS)),.exe,)";\
				env GOOS=$(OS) GOARCH=$(ARCH)\
				go build\
				-o $(OUTPUT_DIR)/$(PACKAGE)/$(APP)_$(PACKAGE)_$(OS)-$(ARCH)$(if $(findstring windows, $(OS)),.exe,)\
				$(PACKAGE_DIR)/$(PACKAGE);\
			)\
		)\
	)

	@echo "\033[0;32m[Finished]\033[0m Output in: $(OUTPUT_DIR)/"

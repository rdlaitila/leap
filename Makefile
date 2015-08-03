#############################################
# Begin Platform Detection
#############################################
ifdef SystemRoot
	BUILD = powershell -ExecutionPolicy bypass -f .\tools\windows\Makefile.ps1 -build
	CLEAN = powershell -ExecutionPolicy bypass -f .\tools\windows\Makefile.ps1 -clean
	INSTALL = powershell -ExecutionPolicy bypass -f .\tools\windows\Makefile.ps1 -install
else
	ifeq ($(shell uname), Linux)
		BUILD = python ./tools/linux/Makefile.py --build
		CLEAN = python ./tools/linux/Makefile.py --clean
		INSTALL = python./tools/linux/Makefile.py --install
	endif
endif

#############################################
# Begin Target Definitions
#############################################
all:
	$(BUILD)

install:
	$(INSTALL)

clean:
	$(CLEAN)

build:
	$(BUILD)

babies:
	@echo "You will need more than a makefile for that..."

.PHONY: all install clean build babies
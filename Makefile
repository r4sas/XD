REPO := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
LOGOS = $(REPO)/contrib/logos
WEBUI_DIR = contrib/webui
WEBUI = ./$(WEBUI_DIR)
GO_ASSETS = $(REPO)/build-assets
DOCROOT = $(WEBUI)/docroot
WEBUI_LOGO = $(DOCROOT)/favicon.png
WEB_FILES = $(DOCROOT)/index.html
WEB_FILES += $(DOCROOT)/xd.min.js
WEB_FILES += $(DOCROOT)/xd.css
WEB_FILES += $(WEBUI_LOGO)
WEBUI_CORE  = $(DOCROOT)/xd.min.js
WEBUI_CORE += $(DOCROOT)/xd.css
WEBUI_PREFIX = /contrib/webui/docroot
ASSETS = $(REPO)/lib/rpc/assets/assets.go

TAGS ?= webui
LOKINET ?= 0
ifeq ($(LOKINET),1)
	TAGS += lokinet
endif

MKDIR = mkdir -p
RM = rm -f
CP = cp
CPLINK = cp -P
INSTALL = install
LINK = ln -s
CHMOD = chmod 

GIT_VERSION ?= $(shell test -e .git && git rev-parse --short HEAD || true)

ifdef GOROOT
	GO = $(GOROOT)/bin/go
else
	GO = $(shell which go)
endif

ifeq ($(GOOS),windows)
	XD := XD.exe
	CLI := XD-cli.exe
	PREFIX ?= /usr/local # FIXME
else
	XD := XD
	CLI := XD-cli
	PREFIX ?= /usr/local
endif

build: $(CLI)


$(XD): webui
	$(GO) build -a -ldflags "-X xd/lib/version.Git=$(GIT_VERSION)" -tags='$(TAGS)' -o $(XD)

dev: webui
	$(GO) build -race -v -a -ldflags "-X xd/lib/version.Git=$(GIT_VERSION)" -tags='$(TAGS)' -o $(XD)

$(CLI): $(XD)
	$(RM) $(CLI)
	$(LINK) $(XD) $(CLI)
	$(CHMOD) 755 $(CLI)

test:
	$(GO) test ./...

clean: webui-clean go-clean
	$(RM) $(CLI)

distclean: clean clean-assets

clean-assets:
	$(RM) $(ASSETS)

webui-clean:
	$(RM) $(WEBUI_LOGO)
	$(MAKE) -C $(WEBUI) clean

go-clean:
	$(GO) clean

$(WEBUI_LOGO):
	$(CP) $(LOGOS)/xd_logo.png $(WEBUI_LOGO)

$(WEBUI_CORE): $(WEBUI_LOGO)
	$(MAKE) -C $(WEBUI)

webui: $(WEBUI_CORE)
	$(CP) $(WEB_FILES) $(REPO)/lib/rpc/assets/

no-webui:
	$(GO) build -ldflags "-X xd/lib/version.Git=$(GIT_VERSION)" -o $(XD)

install: $(XD) $(CLI)
	$(MKDIR) $(PREFIX)/bin
	$(INSTALL) XD $(PREFIX)/bin
	$(CPLINK) $(CLI) $(PREFIX)/bin

LATEST_TAG=$(shell git describe --tags | cut -c 2-)
dist:
	git archive --format=tar.gz -9 --worktree-attributes \
		--prefix=xd_$(LATEST_TAG)/ v$(LATEST_TAG) -o xd_$(LATEST_TAG).orig.tar.gz

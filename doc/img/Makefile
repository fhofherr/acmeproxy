DOCKER := $(shell command -v docker 2> /dev/null)
PODMAN := $(shell command -v podman 2> /dev/null)

PLANTUML_IMAGE := fhofherr/plantuml:1.2019.11
ifdef PODMAN
	PLANTUML := $(PODMAN) run --rm -i $(PLANTUML_IMAGE) -pipe
else ifdef DOCKER
	PLANTUML := $(DOCKER) run --rm -i $(PLANTUML_IMAGE) -pipe
else
	$(error "Neither podman nor docker found.")
endif

SRC_FILES := $(wildcard *.puml)
SVG_FILES := $(patsubst %.puml,%.svg,$(SRC_FILES))
PNG_FILES := $(patsubst %.puml,%.png,$(SRC_FILES))

.DEFAULT_GOAL := svg

%.svg: %.puml
	cat $< | $(PLANTUML) -tsvg > $@

%.png: %.puml
	cat $< | $(PLANTUML) -tpng  > $@

.PHONY: svg
svg: $(SVG_FILES)

.PHONY: png
png: $(PNG_FILES)

.PHONY: clean
clean:
	rm -f *.png
	rm -f *.svg

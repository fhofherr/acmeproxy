PLANTUML := plantuml

SRC_FILES := $(wildcard *.puml)
SVG_FILES := $(patsubst %.puml,%.svg,$(SRC_FILES))
PNG_FILES := $(patsubst %.puml,%.png,$(SRC_FILES))

.DEFAULT_GOAL := svg

%.svg: %.puml
	$(PLANTUML) -tsvg $<

%.png: %.puml
	$(PLANTUML) -tpng $<

.PHONY: svg
svg: $(SVG_FILES)

.PHONY: png
png: $(PNG_FILES)

.PHONY: clean
clean:
	rm -f *.png
	rm -f *.svg

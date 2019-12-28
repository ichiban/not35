BUNDLED = not35
ZIP = assets.zip
RAW = $(BUNDLED).raw
MANIFEST = assets/public/manifest.json

.PHONY: clean run

$(BUNDLED): $(RAW) $(ZIP)
	cat $^ > $@
	zip -A $@
	chmod +x $@

$(ZIP): $(MANIFEST)
	cd assets && zip -r ../$@ .

$(MANIFEST): $(wildcard js/*) $(wildcard css/*)
	webpack

$(RAW): $(wildcard app/*) $(wildcard handlers/*) $(wildcard models/*) $(wildcard views/*)
	go build -o $@

clean:
	rm -f $(BUNDLED) $(ZIP) $(RAW)
	[ -e $(MANIFEST) ] && rm -f $$(sed -n 's|.*".*": "\(.*\)".*|assets/public/\1|p' $(MANIFEST)) || true
	rm -f $(MANIFEST)

run: $(MANIFEST)
	BIND=:8080 DATABASE=postgres://localhost/not35?sslmode=disable SECRET=a454e9cf19a517fcbe4c8f0650f8335c go run ./...

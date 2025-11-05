.PHONY: build serve clean

build:
	cd introspection-ui && pnpm build
	cp introspection-ui/build/index.html spec_ui.html

serve:
	cd introspection-ui && pnpm dev

clean:
	rm -f spec_ui.html
	rm -rf introspection-ui/build
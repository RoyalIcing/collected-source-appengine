build_frontend:
	@frontend/node_modules/.bin/parcel build --out-dir main/public --public-url /public --cache-dir frontend/.cache frontend/index.html
	@echo "\nBuild assets HTML:\n"
	@cat main/public/index.html
	@echo "\n"

dev:
	cd main && make dev

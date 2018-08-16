# https://cloud.google.com/appengine/docs/standard/python/tools/using-local-server

dev:
	dev_appserver.py --support_datastore_emulator=true app.yaml

deploy:
	gcloud app deploy app.prod.yaml --project "${PROJECT}"

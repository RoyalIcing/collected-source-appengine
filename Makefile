# https://cloud.google.com/appengine/docs/standard/python/tools/using-local-server

dev:
	dev_appserver.py --support_datastore_emulator=true app.yaml

dev_reset:
	dev_appserver.py --support_datastore_emulator=true --clear_datastore=yes app.yaml

deploy:
	gcloud app deploy app.prod.yaml --project "${PROJECT}"

# https://cloud.google.com/appengine/docs/standard/python/tools/using-local-server

dev:
	dev_appserver.py --support_datastore_emulator=true --default_gcs_bucket_name "${DEFAULT_BUCKET_NAME}" app.yaml

dev_reset:
	dev_appserver.py --support_datastore_emulator=true --clear_datastore=yes app.yaml

deploy:
	gcloud datastore indexes create index.yaml
	gcloud app deploy app.prod.yaml --project "${PROJECT}"

deploy2:
	python2 /usr/local/Caskroom/google-cloud-sdk/latest/google-cloud-sdk/platform/google_appengine/goapp deploy -application "${PROJECT}" app.prod.yaml

test_integration:
	cd tests-integration && \
	yarn test

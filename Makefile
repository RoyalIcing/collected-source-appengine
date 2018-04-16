dev:
	dev_appserver.py app.yaml

deploy:
	gcloud app deploy app.prod.yaml --project "${PROJECT}"

.PHONY: build deploy clean all all-swap

build:
	GOOS=linux GOARCH=amd64 go build -o ./ocr-service/bootstrap ./ocr-service/cmd/main.go  
	GOOS=linux GOARCH=amd64 go build -o ./vrc-service/bootstrap ./vrc-service/cmd/main.go  
	GOOS=linux GOARCH=amd64 go build -o ./reg_renewal_reminder-service/bootstrap ./reg_renewal_reminder-service/cmd/main.go  
	GOOS=linux GOARCH=amd64 go build -o ./reg_expiration_job-service/bootstrap ./reg_expiration_job-service/cmd/main.go
	cd websocket && zip -r ../deploy-scripts/zip/connect.zip connect
	cd websocket && zip -r ../deploy-scripts/zip/disconnect.zip disconnect
	cd websocket && zip -r ../deploy-scripts/zip/report-authority.zip report-authority

deploy:
	cd deploy-scripts && cdk deploy

deploy-swap:
	cd deploy-scripts && cdk deploy --hotswap

clean:
	rm -rf ./ocr-service/bootstrap
	rm -rf ./reg_renewal_reminder-service/bootstrap
	rm -rf ./vrc-service/bootstrap
	rm -rf ./reg_expiration_job-service/bootstrap
	rm -rf ./deploy-scripts/zip/connect.zip
	rm -rf ./deploy-scripts/zip/disconnect.zip
	rm -rf ./deploy-scripts/zip/report-authority.zip
	
all:
	make clean
	make build
	make deploy

all-swap:
	make clean
	make build
	make deploy-swap
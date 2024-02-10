profile=personal

.PHONY: build clean deploy gomodgen

list=authorizer connection schedule
build:
	for i in $(list); do \
		env GOARCH=arm64 GOOS=linux go build -tags lambda.norpc -o "./bin/$$i/bootstrap" "./cmd/$$i" \
		&& \
		zip -j "./bin/$$i/$$i.zip" "./bin/$$i/bootstrap"; \
		echo "Built $$i"; \
	done

deploy: build
	
	sls deploy --verbose --aws-profile "$(profile)"

gomodgen:
	chmod u+x gomod.sh
	./gomod.sh

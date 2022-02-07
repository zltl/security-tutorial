.PHONY: x docker clean

.ONESHELL:

x:
	go build ./cmd/...
	echo "http server listening on port 8080"
	./st001

docker:
	docker build -t st001 .

run:
	echo "http server listening on port 8080"
	docker run -it --rm -p 8080:8080 st001

clean:
	docker image rm st001
	docker image ls | grep -e "^<none" | awk '{print $3}' | xargs docker image rm -
	rm -rf st001

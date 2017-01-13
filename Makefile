build:
	docker build -t 'creack/gofour:dev' .
	docker run --rm 'creack/gofour:dev' tar -zcf - app /etc/ssl /usr/local/go/lib/time/zoneinfo.zip > .archive.tar.gz

release:
	docker build -t 'creack/gofour:latest' -f Dockerfile.release .

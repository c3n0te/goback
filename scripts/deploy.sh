docker build -f Containerfile --rm -t goback:v0.1 .
docker run -d --rm --name goback goback:v0.1

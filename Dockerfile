FROM golang:1-alpine3.23 AS builder

RUN apk update && apk add build-base

WORKDIR /build

COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY . .
RUN go build -o app .

FROM mysql:8

COPY --from=builder /build/app /usr/local/bin/autodump

### Credits
### https://github.com/docker-library/mysql/blob/f59266f3ec6f1ddfc66d1db613430d9dcc52419b/8.4/Dockerfile.oracle
VOLUME /var/lib/mysql

COPY docker-entrypoint.sh /usr/local/bin/
ENTRYPOINT ["docker-entrypoint.sh"]

EXPOSE 3306 33060
CMD ["mysqld"]

### BUILD ###
# export x_docker_http_proxy="http://host.docker.internal:1080"
# export x_docker_image_name="allape/autodumpmysql:8"
# export x_docker_registry_prefix="docker-registry.lan.allape.cc/"
# export x_docker_registry_image_name="$x_docker_registry_prefix$x_docker_image_name"

# docker build --platform linux/arm64 --build-arg http_proxy=$x_docker_http_proxy --build-arg https_proxy=$x_docker_http_proxy -f Dockerfile -t "$x_docker_image_name-arm64" .
# docker build --platform linux/amd64 --build-arg http_proxy=$x_docker_http_proxy --build-arg https_proxy=$x_docker_http_proxy -f Dockerfile -t $x_docker_image_name .
# docker tag $x_docker_image_name $x_docker_registry_image_name && docker push $x_docker_registry_image_name

# sudo docker pull $x_docker_registry_image_name && sudo docker tag $x_docker_registry_image_name $x_docker_image_name
# sudo docker run -d --name mysqltest -p 3306:3306 -e "MYSQL_ROOT_PASSWORD=Root_123456" -v "$(pwd)/database:/var/lib/mysql" $x_docker_image_name

## Testing autodump with `--dumpnow` must be with user `mysql` instead of `root`.
## If running test is before a cron-triggered dump, there will/might occur a `permission denied` error.
# sudo docker exec mysqltest gosu mysql autodump --dumpnow

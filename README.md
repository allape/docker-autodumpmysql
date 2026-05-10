# docker-autodumpmysql
Docker build for MySQL with auto dump.

## Usage

### Docker Compose

```yaml
name: demo
services:
  app:
    image: alpine:3
    command:
      - tail
      - -F
      - /dev/void
    restart: unless-stopped
    networks:
      - internal
      - gate
    depends_on:
      database:
        condition: service_healthy
  database:
    image: allape/autodumpmysql:8
    restart: unless-stopped
    networks:
      - internal
    healthcheck:
      test: mysql -uroot -p"$$MYSQL_ROOT_PASSWORD" -e "SELECT 1" > /dev/null 2>&1 || exit 1
    environment:
      - MYSQL_ROOT_PASSWORD=rIHdF5gUKmyAFvdtgUYL
      - MYSQL_DATABASE=demo
    volumes:
      - ./database:/var/lib/mysql
networks:
  gate:
  internal:
    internal: true
```

### Docker

```shell
sudo docker run -d --name mysqltest -p 3306:3306 -e "MYSQL_ROOT_PASSWORD=Root_123456" -v "$(pwd)/database:/var/lib/mysql" allape/autodumpmysql:8
sudo docker exec mysqltest gosu mysql autodump --dumpnow
```

## Build

See [Dockerfile](Dockerfile) for detail.

# Credits
MySQL: https://github.com/docker-library/mysql

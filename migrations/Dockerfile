FROM alpine:3.23

WORKDIR /migrate

RUN apk add --no-cache curl

RUN curl -fsSL \
        https://raw.githubusercontent.com/pressly/goose/master/install.sh |\
        sh -s v3.27.1

COPY migrate.sh ./
RUN chmod +x ./migrate.sh

COPY *.sql ./

CMD ./migrate.sh

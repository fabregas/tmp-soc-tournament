FROM golang:1.8-alpine

RUN apk --update add postgresql-client git && rm -rf /var/cache/apk/*

WORKDIR /go/src/app

COPY . .

RUN go get github.com/lib/pq && go build -o test-app .

RUN ls -la
CMD sleep 10 && psql -U postgres -h pg -d test_tournament < schema.sql && \ 
    ./test-app -db-conn postgres://postgres@pg/test_tournament?sslmode=disable 


FROM golang:1.20-buster

WORKDIR /banner-service/server
RUN mkdir controllers db middlewares schemas

COPY server/controllers/ controllers/
COPY server/db/ db/
COPY server/middlewares/ middlewares/
COPY server/schemas/ schemas/
COPY main.go main.go

COPY go.mod go.mod
RUN go mod tidy
RUN go build || exit 1

ENTRYPOINT [ "./banner-service" ]
FROM golang:1.20-buster

WORKDIR /bannerservice/server
RUN mkdir controllers db routes middlewares schemas

COPY controllers/ controllers/
COPY db/ db/
COPY middlewares/ middlewares/
COPY routes/ routes/
COPY schemas/ schemas/
COPY main.go main.go

COPY go.mod go.mod
RUN go mod tidy
RUN go build || exit 1

ENTRYPOINT [ "./server" ]
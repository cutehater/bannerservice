FROM golang:1.20-buster

WORKDIR /bannerservice/server
RUN mkdir controllers db routes middlewares schemas

COPY ../../controllers/ controllers/
COPY ../../db/ db/
COPY ../../routes/ routes/
COPY ../../middlewares/ middlewares/
COPY ../../schemas/ schemas/
COPY . test/

COPY go.mod go.mod
RUN go mod tidy
RUN go build || exit 1

CMD [ "go", "test" ]
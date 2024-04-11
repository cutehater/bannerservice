FROM golang:1.20-buster

WORKDIR /bannerservice/server/test
RUN mkdir controllers db routes middlewares schemas

COPY controllers/ controllers/
COPY db/ db/
COPY routes/ routes/
COPY middlewares/ middlewares/
COPY schemas/ schemas/
COPY test/integration/ test/

COPY ../go.mod go.mod
RUN go mod tidy

CMD [ "go", "test", "-v", "./test/..."]
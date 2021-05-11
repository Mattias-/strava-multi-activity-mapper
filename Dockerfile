FROM golang:1.16
COPY . /src
WORKDIR /src

RUN echo "commit="$(git describe --always --dirty)"">>envfile
RUN echo "date="$(git show -s --format="%cI")"">>envfile

RUN . ./envfile; CGO_ENABLED=0 go build -ldflags "-s -w -X main.commit=$commit -X main.date=$date" ./cmd/mam/

FROM scratch
COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=0 /src/mam /mam
COPY ./static /static
CMD ["/mam"]

FROM golang:1.14.2
COPY . /src
WORKDIR /src
RUN CGO_ENABLED=0 go build -o mam

FROM scratch
COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=0 /src/mam /mam
COPY ./static /static
CMD ["/mam"]

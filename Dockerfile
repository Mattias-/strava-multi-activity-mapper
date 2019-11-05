FROM golang:1.13.4
COPY . /src
WORKDIR /src
RUN CGO_ENABLED=0 go build -o mam

FROM scratch
COPY --from=0 /src/mam /mam
COPY ./index.html /index.html
COPY ./static /static
CMD ["/mam"]

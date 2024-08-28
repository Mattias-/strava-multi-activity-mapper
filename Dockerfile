FROM node:20 AS frontend
COPY frontend /src/frontend
WORKDIR /src/frontend
RUN npm ci && npm run build

FROM golang:1.23 AS backend
COPY . /src
WORKDIR /src
RUN CGO_ENABLED=0 go build -ldflags "-s -w" ./cmd/mam/

FROM scratch
COPY --from=backend /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=backend /src/mam /mam
COPY --from=frontend /src/frontend/dist /frontend/dist
CMD ["/mam"]

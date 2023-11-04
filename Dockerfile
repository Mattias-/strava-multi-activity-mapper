FROM node:16 AS frontend
COPY frontend /src/frontend
WORKDIR /src/frontend
RUN npm ci && npm run build

FROM golang:1.21 AS backend
COPY . /src
WORKDIR /src

RUN echo "commit="$(git describe --always --dirty)"">>envfile
RUN echo "date="$(git show -s --format="%cI")"">>envfile

RUN . ./envfile; CGO_ENABLED=0 go build -ldflags "-s -w -X main.commit=$commit -X main.date=$date" ./cmd/mam/

FROM scratch
COPY --from=backend /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=backend /src/mam /mam
COPY --from=frontend /src/frontend/dist /frontend/dist
CMD ["/mam"]

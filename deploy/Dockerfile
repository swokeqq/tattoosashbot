FROM golang:1.27-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
COPY migrations ./migrations

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /bot ./cmd/bot

FROM gcr.io/distroless/static-debian13:nonroot

COPY --from=build /bot /bot

ENTRYPOINT ["/bot"]
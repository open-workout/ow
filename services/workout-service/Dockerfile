FROM golang:1.25-alpine AS builder
WORKDIR /workspace
COPY go.work go.work.sum ./
# go work sync requires all workspace module go.mod files to be present
COPY shared/contracts/go.mod ./shared/contracts/
COPY shared/env/go.mod ./shared/env/
COPY services/api-gateway/go.mod services/api-gateway/go.sum ./services/api-gateway/
COPY services/user-service/go.mod services/user-service/go.sum ./services/user-service/
COPY services/exercise-service/go.mod services/exercise-service/go.sum ./services/exercise-service/
COPY services/workout-service/go.mod services/workout-service/go.sum ./services/workout-service/
RUN go work sync
COPY shared/ ./shared/
COPY services/workout-service/ ./services/workout-service/
RUN CGO_ENABLED=0 go build -o /app ./services/workout-service/cmd/...

FROM alpine:3.19
COPY --from=builder /app /app
EXPOSE 8082
ENTRYPOINT ["/app"]

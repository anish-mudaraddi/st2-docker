# Build stage
FROM golang:1.22.4-alpine AS build

WORKDIR /app

COPY auth_service/go.mod .
RUN go mod download

COPY auth_service/*.go .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o auth_service .

# Final stage
FROM nginx:alpine

COPY --from=build /app/auth_service /app/auth_service
COPY nginx.conf /etc/nginx/nginx.conf
COPY start.sh /start.sh

RUN chmod +x /start.sh

EXPOSE 80

CMD ["/start.sh"]
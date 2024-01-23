FROM golang:1.20
WORKDIR /identity-forecaster
COPY go.mod go.sum ./
RUN go mod download
COPY . /identity-forecaster
RUN CGO_ENABLED=0 GOOS=linux go build -o /identity-forecaster/cmd/forecaster/bin/main /identity-forecaster/cmd/forecaster/main.go
CMD ["bash", "-c", "/identity-forecaster/cmd/forecaster/bin/main"]
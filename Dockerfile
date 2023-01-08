FROM olivercvetkovski/zvax-protoc:latest AS protoc
FROM golang:1.19 AS build

WORKDIR /app

# copy proto files
COPY --from=protoc /app/common ./common

WORKDIR /app/src

# copy go mod and sum files
COPY go.mod .
COPY go.sum .

# download dependencies
RUN go mod download

# copy source code
COPY . .

# build the Go app
RUN go build -o /run/zvax-auth .

# final stage
FROM scratch AS final

WORKDIR /app

# copy the executable
COPY --from=build /run/zvax-auth .

# run the executable
ENTRYPOINT [ "./zvax-auth" ]

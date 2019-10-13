FROM golang:1.13.1 as builder

LABEL maintainer="Isaac Garza <garzai@alum.mit.edu>"

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .


FROM alpine:latest  

RUN echo "http://dl-cdn.alpinelinux.org/alpine/edge/testing" >> /etc/apk/repositories

RUN apk --no-cache add ca-certificates fluidsynth ffmpeg
# RUN apk update && apk add fluidsynth ffmpeg

WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/main .

CMD [ "./main" ]
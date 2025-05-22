FROM golang:1.24-bullseye

RUN apt-get update && apt-get install -y \
    libsdl2-dev \
    libsdl2-ttf-dev \
    libsdl2-image-dev \
    libsdl2-gfx-dev

WORKDIR /build

COPY go.mod go.sum* ./

RUN GOWORK=off go mod download

COPY . .
RUN GOWORK=off go build -gcflags="all=-N -l" -v

CMD ["/bin/bash"]
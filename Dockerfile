FROM --platform=linux/arm64 golang:1.24-bullseye

RUN apt-get update && apt-get install -y \
    libsdl2-dev \
    libsdl2-ttf-dev \
    libsdl2-image-dev

WORKDIR /build

COPY go.mod go.sum* ./

RUN go mod download

COPY . .
RUN go build -gcflags="all=-N -l" -v

CMD ["/bin/bash"]
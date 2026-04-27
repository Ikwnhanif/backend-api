# Tahap 1: Build (Menggunakan image Go lengkap)
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o main .

# Tahap 2: Runner (Hanya mengambil hasil build, sangat ringan)
FROM alpine:latest
WORKDIR /app
# Install tzdata agar zona waktu WIB/Jakarta akurat
RUN apk add --no-cache tzdata
ENV TZ=Asia/Jakarta

COPY --from=builder /app/main .
# Pastikan Anda punya file .env-production nanti
COPY .env . 

EXPOSE 3001
CMD ["./main"]
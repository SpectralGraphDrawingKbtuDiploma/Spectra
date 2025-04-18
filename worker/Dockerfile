# Stage 1: Сборка Go-приложения
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Устанавливаем зависимости для Go
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Компиляция Go-приложения
RUN go build -o /app/server ./cmd

# Stage 2: Финальный образ с Python, C++ и Eigen
FROM ubuntu:latest

WORKDIR /app

# Устанавливаем необходимые пакеты (используем apt вместо apk)
RUN apt update && apt install -y \
    python3 python3-pip python3-venv \
    g++ cmake wget \
    git \
    && rm -rf /var/lib/apt/lists/*

# Создаём виртуальное окружение и устанавливаем Python-зависимости
RUN python3 -m venv /app/venv \
    && /app/venv/bin/pip install --no-cache-dir numpy matplotlib boto3

# Устанавливаем Eigen
ARG EIGEN_PATH=/usr/local/include/eigen3/
RUN mkdir -p /usr/local/include/eigen3 && \
    wget -qO- https://gitlab.com/libeigen/eigen/-/archive/3.4.0/eigen-3.4.0.tar.gz | tar xz && \
    cp -r eigen-3.4.0/Eigen $EIGEN_PATH

# Устанавливаем Spectra как header-only библиотеку
ARG SPECTRA_PATH=/usr/local/include/spectra/
RUN git clone --branch master https://github.com/yixuan/spectra.git /tmp/spectra && \
    cp -r /tmp/spectra/include $SPECTRA_PATH && \
    rm -rf /tmp/spectra

# Копируем сервер из первого этапа
COPY --from=builder /app/server /app/server

# Копируем скрипты
COPY ./draw.py .
COPY ./draw.sh .
COPY ./script.cpp .
COPY ./graph.txt ./graph/graph.txt
COPY ./draw.c .
COPY ./stb_image_write.h .
COPY ./upload_to_s3.py .
COPY ./cleaner.py .
COPY ./eigen3D.cpp .
COPY ./gen_obj.py .

# Открываем порт
EXPOSE 8080

# Запускаем сервер
CMD ["/app/server"]
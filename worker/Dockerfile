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
    && rm -rf /var/lib/apt/lists/*

# Создаём виртуальное окружение и устанавливаем Python-зависимости
RUN python3 -m venv /app/venv \
    && /app/venv/bin/pip install --no-cache-dir numpy matplotlib


ARG EIGEN_PATH=/usr/local/include/eigen3/
# Клонируем и устанавливаем Eigen
RUN mkdir -p /usr/local/include/eigen3 && \
    wget -qO- https://gitlab.com/libeigen/eigen/-/archive/3.4.0/eigen-3.4.0.tar.gz | tar xz && \
    cp -r eigen-3.4.0/Eigen $EIGEN_PATH

# Копируем сервер из первого этапа
COPY --from=builder /app/server /app/server

# Копируем скрипты
COPY ./draw.py .
COPY ./draw.sh .
COPY ./script.cpp .
COPY ./graph.txt ./graph/graph.txt
COPY ./draw.c .
COPY ./stb_image_write.h .

# Открываем порт
EXPOSE 8080

# Запускаем сервер
CMD ["/app/server"]

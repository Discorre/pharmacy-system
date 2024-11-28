FROM golang:1.20-alpine

# Добавляем репозитории
RUN echo "http://dl-cdn.alpinelinux.org/alpine/v3.14/main" >> /etc/apk/repositories
RUN echo "http://dl-cdn.alpinelinux.org/alpine/v3.14/community" >> /etc/apk/repositories

# Устанавливаем bash
RUN apk update && apk add --no-cache bash

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем файлы проекта в контейнер
COPY . .

# Устанавливаем зависимости
RUN go mod download

# Собираем приложение
RUN go build -o main .

# Открываем порт 8080
EXPOSE 8080

# Запускаем приложение
CMD ["./main" ,"/bin/bash"]
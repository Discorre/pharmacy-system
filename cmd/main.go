package main

import (
    "log"
    "net/http"
    "pharmacy-system/internal/api"
    "pharmacy-system/internal/db"
    "pharmacy-system/internal/metrics"

    "github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
    // Инициализация метрик
    metrics.InitMetrics()

    // Подключение к базе данных
    conn, err := db.ConnectDB()
    if err != nil {
        log.Fatalf("Ошибка подключения к базе данных: %v", err)
    }
    defer conn.Close()

    // Инициализация маршрутов
    router := api.NewRouter(conn)

    // Добавление middleware для метрик
    wrappedRouter := metrics.MetricsMiddleware(router)

    // Экспорт метрик Prometheus
    http.Handle("/metrics", promhttp.Handler())

    // Запуск сервера
    log.Println("Сервер запущен на :8080")
    log.Fatal(http.ListenAndServe(":8080", wrappedRouter))
}

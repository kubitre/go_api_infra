# Go Infra Api

Infrastructre methods and types for make api simple

## Response

- abstract type ApiResponse which can used as Tipically responses code as functions
- Json Style for ApiError and method for simple constructing this model

## Request

- extract body to data class from all sources request with special tag

## Metrics Prometheus

- base golden metrics for api with high level methods for sending metrics to prometheus
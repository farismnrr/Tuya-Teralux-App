package com.example.teraluxapp.data.model

data class SensorDataResponse(
    val temperature: Double,
    val humidity: Int,
    val battery_percentage: Int,
    val status_text: String,
    val temp_unit: String
)

package com.example.teraluxapp.data.model

import com.google.gson.annotations.SerializedName

data class Device(
    @SerializedName("id") val id: String,
    @SerializedName("remote_id") val remoteId: String?,  // For IR devices
    @SerializedName("name") val name: String,
    @SerializedName("category") val category: String,
    @SerializedName("product_name") val productName: String,
    @SerializedName("online") val online: Boolean,
    @SerializedName("icon") val icon: String,
    @SerializedName("status") val status: List<DeviceStatus>?,
    @SerializedName("ip") val ip: String?,
    @SerializedName("local_key") val localKey: String?,
    @SerializedName("gateway_id") val gatewayId: String?,
    @SerializedName("collections") val collections: List<Device>?
)

data class DeviceStatus(
    @SerializedName("code") val code: String,
    @SerializedName("value") val value: Any? // Any? for generic JSON value
)

data class DeviceResponse(

    @SerializedName("devices") val devices: List<Device>,
    @SerializedName("total_devices") val totalDevices: Int,
    @SerializedName("current_page_count") val currentPageCount: Int
)

data class SingleDeviceResponse(
     @SerializedName("device") val device: Device
)

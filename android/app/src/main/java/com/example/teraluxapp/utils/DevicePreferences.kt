package com.example.teraluxapp.utils

import android.content.Context
import android.content.SharedPreferences

class DevicePreferences(context: Context) {
    private val prefs: SharedPreferences = context.getSharedPreferences("device_prefs", Context.MODE_PRIVATE)

    fun saveACState(deviceId: String, isOn: Boolean, temp: Int, mode: Int, speed: Int) {
        with(prefs.edit()) {
            putBoolean("${deviceId}_ison", isOn)
            putInt("${deviceId}_temp", temp)
            putInt("${deviceId}_mode", mode)
            putInt("${deviceId}_speed", speed)
            apply()
        }
        android.util.Log.d("DevicePreferences", "Saved state for $deviceId: On=$isOn, Temp=$temp, Mode=$mode")
    }

    fun getACState(deviceId: String): ACState {
        val state = ACState(
            isOn = prefs.getBoolean("${deviceId}_ison", false),
            temp = prefs.getInt("${deviceId}_temp", 24),
            mode = prefs.getInt("${deviceId}_mode", 0),
            speed = prefs.getInt("${deviceId}_speed", 0)
        )
        android.util.Log.d("DevicePreferences", "Loaded state for $deviceId: $state")
        return state
    }

    fun saveSwitchState(deviceId: String, switch1: Boolean, switch2: Boolean) {
        with(prefs.edit()) {
            putBoolean("${deviceId}_sw1", switch1)
            putBoolean("${deviceId}_sw2", switch2)
            apply()
        }
        android.util.Log.d("DevicePreferences", "Saved switch state for $deviceId: Sw1=$switch1, Sw2=$switch2")
    }

    fun getSwitchState(deviceId: String): SwitchState {
        val state = SwitchState(
            switch1 = prefs.getBoolean("${deviceId}_sw1", false),
            switch2 = prefs.getBoolean("${deviceId}_sw2", false)
        )
        android.util.Log.d("DevicePreferences", "Loaded switch state for $deviceId: $state")
        return state
    }
}

data class SwitchState(
    val switch1: Boolean,
    val switch2: Boolean
)

data class ACState(
    val isOn: Boolean,
    val temp: Int,
    val mode: Int,
    val speed: Int
)

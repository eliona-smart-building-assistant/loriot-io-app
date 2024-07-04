# Loriot.io User Guide

### Introduction

> The Loriot.io app provides integration and synchronization between Eliona and Loriot.io services and LoRaWAN devices.

## Overview

This guide provides instructions on configuring, installing, and using the Loriot.io app to manage resources and synchronize LoRaWAN devices between Eliona and Loriot.io services.

## Installation

Install the Loriot.io app via the Eliona App Store.

## Configuration

The Loriot.io app requires configuration through Eliona’s settings interface. Below are the general steps and details needed to configure the app effectively.

### Registering the app in Loriot.io Service

Create credentials in Loriot.io Service to connect the Loriot.io services from Eliona. All required credentials are listed below in the [configuration section](#configure-the-loriot-io-app).  

To connect the [Loriot.io API](https://docs.loriot.io/space/LNS/6231610/User+API+7.0) you have to ask your provider to get an API-Key.

### Configure the Loriot.io app 

Configurations can be created in Eliona under `Apps > Loriot.io > Settings` which opens the app's [Generic Frontend](https://doc.eliona.io/collection/v/eliona-english/manuals/settings/apps). Here you can use the `/configs` endpoint with the POST method. Each configuration requires the following data:

| Attribute         | Description                                     |
|-------------------|-------------------------------------------------|
| `baseURL`         | URL of the Loriot.io services.                  |
| `api_token`       | API Token to access the API.                    |
| `enable`          | Flag to enable or disable this configuration.   |
| `refreshInterval` | Interval in seconds for data synchronization.   |
| `requestTimeout`  | API query timeout in seconds.                   |
| `projectIDs`      | List of Eliona project IDs for data collection. |

Example configuration JSON:

```json
{
  "baseURL": "http://service/v1",
  "api_token": "53cr3t",
  "enable": true,
  "refreshInterval": 60,
  "requestTimeout": 120,
  "projectIDs": [
    "10"
  ]
}
```

To define devices handled by the Loriot.io app it is necessary to configure these devices. Here you can use the `/devices` endpoint with the POST method. If the device still don't exist it will be registered in Loriot.io as well. 

Example device configuration via OTAA v1.0 in JSON:

```json
{
    "devEUI": "0123456789ABCDEF",
    "appID": "1234ABCD",
    "assetTypeName": "Device",
    "configID": 1,
    "title": "LoRaWAN test device",
    "description": "This is a LoRaWAN test device",
    "appEUI": "1000000000000000",
    "appKey": "secret"
}
```

## Continuous Asset Creation

Once configured and devices created, the app starts Continuous Asset Creation (CAC). Discovered resources are automatically created as assets in Eliona, and users are notified via Eliona’s notification system.

## Additional Features

### Device Update

You can change the title and the description of a device asset in Eliona. These changes are synchronized automatically into Loriot.io.
If you delete an asset in Eliona the corresponding device is unregistered in Loriot.io as well.

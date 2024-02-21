# Eliona app to access Loriot.io to handle LoRaWAN devices

This app connects [Loriot.io](https://www.loriot.io) network using the [API](https://docs.loriot.io/display/LNS/User+API+7.0) and synchronize information about LoRaWAN devices and the corresponding Eliona assets.

## Configuration

The app needs environment variables and database tables for configuration. To edit the database tables the app provides an own API access.

### Registration in Eliona ###

To start and initialize an app in an Eliona environment, the app has to be registered in Eliona. For this, entries in database tables `public.eliona_app` and `public.eliona_store` are necessary.

This initialization can be handled by the `reset.sql` script.

### Environment variables

- `CONNECTION_STRING`: configures the [Eliona database](https://github.com/eliona-smart-building-assistant/go-eliona/tree/main/db). Otherwise, the app can't be initialized and started (e.g. `postgres://user:pass@localhost:5432/iot`).

- `INIT_CONNECTION_STRING`: configures the [Eliona database](https://github.com/eliona-smart-building-assistant/go-eliona/tree/main/db) for app initialization like creating schema and tables (e.g. `postgres://user:pass@localhost:5432/iot`). Default is content of `CONNECTION_STRING`.

- `API_ENDPOINT`:  configures the endpoint to access the [Eliona API v2](https://github.com/eliona-smart-building-assistant/eliona-api). Otherwise, the app can't be initialized and started. (e.g. `http://api-v2:3000/v2`)

- `API_TOKEN`: defines the secret to authenticate the app and access the Eliona API.

- `API_SERVER_PORT`(optional): define the port the API server listens. The default value is Port `3000`.

- `LOG_LEVEL`(optional): defines the minimum level that should be [logged](https://github.com/eliona-smart-building-assistant/go-utils/blob/main/log/README.md). The default level is `info`.

### Database tables ###

The app requires configuration data that remains in the database. To do this, the app creates its own database schema `loriot-io` during initialization. To modify and handle the configuration data the app provides an API access. Have a look at the [API specification](https://eliona-smart-building-assistant.github.io/open-api-docs/?https://raw.githubusercontent.com/eliona-smart-building-assistant/loriot-io-app/develop/openapi.yaml) how the configuration tables should be used.

- `loriot_io.configuration`: Contains configuration of the app. Editable through the API.

- `loriot_io.asset`: Provides asset mapping. Maps LoRaWAN devices to Eliona asset IDs.

## References

### App API ###

The app provides its own API to access configuration data and other functions. The full description of the API is defined in the `openapi.yaml` OpenAPI definition file.

- [API Reference](https://eliona-smart-building-assistant.github.io/open-api-docs/?https://raw.githubusercontent.com/eliona-smart-building-assistant/loriot-io-app/develop/openapi.yaml) shows details of the API

### Configuring the app ###

To use the app it is necessary to create at least one configuration. A configuration points to one Loriot.io network API endpoint.
Also, a configuration defines at least one Eliona project ID for automatic asset creation.

A minimum configuration that can used by the app's API endpoint `POST /configs` is:

```json
{
    "apiBaseUrl": "https://eu1.loriot.io",
    "apiToken": "secret",
    "enable": true,
    "projectIDs": ["10"]
}
```

### Creating new LoRaWAN devices ###

The app can handle the creation of new LoRaWAN devices with the `PUT /devices` endpoint. For devices created with this endpoint
a corresponding asset in Eliona is created for each Eliona project defined in the configuration.

Minimum example to create a new device via OTAA v1.0 is:

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

## Tools

### Generate API server stub ###

For the API server the [OpenAPI Generator](https://openapi-generator.tech/docs/generators/openapi-yaml) for go-server is used to generate a server stub. The easiest way to generate the server files is to use one of the predefined generation script which use the OpenAPI Generator Docker image.

```
.\generate-api-server.cmd # Windows
./generate-api-server.sh # Linux
```

### Generate Database access ###

For the database access [SQLBoiler](https://github.com/volatiletech/sqlboiler) is used. The easiest way to generate the database files is to use one of the predefined generation script which use the SQLBoiler implementation. Please note that the database connection in the `sqlboiler.toml` file have to be configured.

```
.\generate-db.cmd # Windows
./generate-db.sh # Linux
```

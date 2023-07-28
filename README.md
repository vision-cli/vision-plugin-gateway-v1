# ![logo](./images/vision-logo.svg "Vision") &nbsp; Vision Plugin - Gateway

This plugin creates a standard gateway template

Vision plugins require golang (https://go.dev) to be installed

Install the plugin with

```
go install github.com/vision-cli/vision-plugin-gateway-v1
```

You will now see the gateway plugin commands on the vision cli

```
vision --help
```

# vision-plugin-gateway-v1

From the root of your project run the follownig command

```
vision gateway create <gateway name e.g. gateway>
```

This will create a gateway in the services default folder. The gateway is http (not grpc) and is internet facing
and hence must be in the default services namespace by convention.

If you need to rebuild the gateway after adding new service, you can run the same command again and it will
rebuild the gateway.

You must let the pipeline run in order to create the gateway container image.

# config

After the gateway has been deployed, you will need to configure the following environment variables

- GRPC_PORT = 443
  for each service:
- <SERVICE NAMESPACE>\_<SERVICE_NAME>\_HOST = host name as showin in the cloud run console (without the https)
  e.g. NAMESPACE_SERVICE_HOST = service-1-abc-uc.a.run.app

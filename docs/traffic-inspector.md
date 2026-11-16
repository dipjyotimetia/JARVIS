# Traffic inspector

`jarvis proxy` proxies HTTP, and HTTPS when TLS is enabled. It can record traffic, replay it, or pass it through. The web UI listens on port `9090` by default at `http://localhost:9090/ui/`.

Configuration precedence is command-line flags, then environment variables, then `config.yaml`, then defaults.

## Basic usage

```bash
jarvis proxy

jarvis proxy --record

jarvis proxy --replay

jarvis proxy --ui-port=9999
```

The HTTP proxy listens on port `8080` unless `--http-port` is set. `--target-url` is the default upstream when no path route matches.

## HTTPS and mTLS

```bash
jarvis certificate --cert-dir ./certs

jarvis proxy --tls --cert ./certs/server.crt --key ./certs/server.key --tls-port=8443

jarvis proxy --mtls --client-ca ./certs/ca.crt
```

`--insecure` allows insecure TLS connections to target servers. `--client-cert` and `--client-key` are the client certificate and key for outbound mTLS.

## OpenAPI validation

```bash
jarvis proxy --api-validate --api-spec ./specs/api.yaml

jarvis proxy --api-validate --api-spec ./specs/api.yaml --strict-validation
```

`--validate-req` and `--validate-resp` default to true. `--continue-on-error` keeps processing when validation fails.

## Environment variables

```bash
export JARVIS_HTTP_PORT=8080
export JARVIS_UI_PORT=9090
export JARVIS_TLS_ENABLED=true
```

Viper maps `JARVIS_` variables onto config keys, with dots replaced by underscores (`JARVIS_TLS_ENABLED` sets `tls.enabled`).

## Config file

```yaml
http_port: 8080
ui_port: 9090
recording_mode: false

tls:
  enabled: true
  port: 8443
  cert_file: "./certs/server.crt"
  key_file: "./certs/server.key"

api_validation:
  enabled: true
  spec_path: "./specs/api.yaml"
```

Path-based routing uses `target_routes` (`path_prefix` and `target_url`). `http_target_url` is the fallback upstream. More configuration examples are in the [setup guide](setup.md).

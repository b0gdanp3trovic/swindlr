Swindlr is a load balancer designed to distribute traffic efficiently across multiple backend servers. It supports a variety of features including response caching, SSL/TLS termination, dynamic backend management, multiple load balancing strategies, sticky sessions and rate limiting . This document describes the configuration options available for Swindlr.


# Swindlr Load Balancer Configuration

This document outlines the configuration options available for the Swindlr load balancer. The configuration is managed using [Viper](https://github.com/spf13/viper) and can be set via a configuration file, environment variables, or command-line flags.

## Configuration Options

### General Settings

- **port**: The port on which the load balancer will listen for incoming HTTP requests.
  - Default: `8080`
  - Environment Variable: `PORT`

- **backends**: A list of backend server URLs that the load balancer will distribute traffic to.
  - Default: `[]`
  - Environment Variable: `BACKENDS`

### SSL/TLS Settings

- **use_ssl**: Enable or disable SSL/TLS for the load balancer.
  - Default: `false`
  - Environment Variable: `USE_SSL`

- **ssl_cert_file**: Path to the SSL certificate file.
  - Default: `""`
  - Environment Variable: `SSL_CERT_FILE`

- **ssl_key_file**: Path to the SSL key file.
  - Default: `""`
  - Environment Variable: `SSL_KEY_FILE`

### Dynamic Backend Management

- **use_dynamic**: Enable or disable dynamic management of backend servers via API.
  - Default: `false`
  - Environment Variable: `USE_DYNAMIC`

- **apiPort**: The port on which the API server for dynamic backend management will listen.
  - Default: `8082`
  - Environment Variable: `API_PORT`

### Load Balancing Strategy

- **load_balancer.strategy**: The strategy used for load balancing. Valid options are:
  - `round_robin`
  - `least_connections`
  - `random`
  - `latency_aware` - chooses a backend depending on the latency. The latency is recorded periodically.
  - Default: `round_robin`
  - Environment Variable: `LOAD_BALANCER_STRATEGY`

### Sticky Sessions

- **use_sticky_sessions**: Enable or disable sticky sessions, which bind a client to a specific backend server.
  - Default: `false`
  - Environment Variable: `USE_STICKY_SESSIONS`

### Rate Limiting

- **rate_limiting.rate**: The rate at which requests are allowed to pass through the rate limiter (requests per second).
  - Default: `10.0`
  - Environment Variable: `RATE_LIMITING_RATE`

- **rate_limiting.bucket_size**: The maximum number of requests that can be allowed to pass through the rate limiter in a burst.
  - Default: `5`
  - Environment Variable: `RATE_LIMITING_BUCKET_SIZE`

### Caching

- **use_cache**: Enable or disable caching of responses.
  - Default: `false`
  - Environment Variable: `USE_CACHE`

## Example Configuration File

Below is an example `config.yaml` file that sets various configuration options:

```yaml
port: 8080
backends:
  - http://backend1.example.com
  - http://backend2.example.com
use_ssl: true
ssl_cert_file: "/path/to/cert.pem"
ssl_key_file: "/path/to/key.pem"
use_dynamic: true
apiPort: 8082
load_balancer:
  strategy: least_connections
use_sticky_sessions: true
rate_limiting:
  rate: 20.0
  bucket_size: 10
use_cache: true

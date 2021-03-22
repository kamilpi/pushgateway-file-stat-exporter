# Description
A small and simple exporter for files in the specific directory. It is ready to push metrics to [pushgateway](https://github.com/prometheus/pushgateway)
  
# Environment variables
| Name        | Description           | Default  |
| ------------- |:-------------:| -----:|
| PUSHGATEWAY_URL     | Pushgateway endpoint ||
| TLS_SKIP_VERIFY      | Skip verification of valid certificate when pushgateway is exposing via HTTPS      |   0 |
| DIR1_PATH | Path to directory where files should be collected      ||
|DIR1_LABEL|Label of a directory. It will be useful for specific purpose, e.g. presentation in the Grafana||
|DIR1_EXT|Extension of files which should be collected||

# Contributing  
When contributing to this repository, please first discuss the change you wish to make via issue, email, or any other method with the owners of this repository before making a change.
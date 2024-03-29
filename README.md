# PiSec-Brain
Brain is the main server of the PiSec project (https://github.com/ringloop/pisec)

The aim of this project is to create a server containing a huge amount of URLs marked as malicious (phishing, malware, etc) and expose services to other PiSec components.

To maximize efficiency in URL search, an ElasticSearch database (https://www.elastic.co/) is used to store the data. This makes data insertion a bit more expensive, but the search is highly efficient. 

Services to interact with other PiSec components are implemented as REST endpoints, in particular:

- /api/v1/indicator/url: is the endpoint exposed fro crawlers and other clients providing new phishing URLs. This endpoint expects a JSON containing a list of input data. Detaled structure definition can be found in model.go source file

- /api/v1/indicators: This endpoint is used by the Proxy, downloading the entire set of phishing URLs known by the server, in the format of a single Bloom Filter. This data structure allowed us to express a huge amount of informations in a tiny data structure. The Wikipedia page of Bloom Filters (https://en.wikipedia.org/wiki/Bloom_filter) provides detailed informations.

- /api/v1/indicators/details: This endpoint allows the check of a single URL. This service is used by the proxy to check the real presence of the provided URL in the database. This service is useful due to the probabilistic nature of Bloom Filter, in case of a positive match, the URL must be checked. Caching mechanisms are used by Pisec Proxy to avoid request flooding.

At the moment, the configuration is taken from the following environment variables:
  - ES_HOST: the ElasticSearch host address (example: http://localhost:9200) 
  - ES_USER: the configured ElasticSearch instance username
  - ES_PASSWD: the configured ElasticSearch instance password

The server can be started as a common go process (with go run cmd/main.go) or in a docker instance (using the provided Dockerfile configuring the proper environment variables mentioned above). 




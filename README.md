# go-core
Commonly used functions in my personal projects.

This includes:
* [Logging](./logging) setup for slog with environment variables
* A [basic layer](./app) for dependency injection and app state management
* [A package](./env) for parsing environment variables
* [A server](./httpx) that wraps over [chi](https://github.com/go-chi/chi) to make the initialisation and closing of such a server more easy to work with

Planned to do later when needed: 
* Tracing with OTEL (https://www.jaegertracing.io/)
  * https://medium.com/jaegertracing/experiment-migrating-opentracing-based-application-in-go-to-use-the-opentelemetry-sdk-29b09fe2fbc4

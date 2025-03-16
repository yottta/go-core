# reno-core
The core functionality for reno

This includes:
* Data fetcher API
  * This is defining the interface of a data fetcher. The data fetcher itself can be anything, from a TMBD entry, to a github repo or even also a particular tool that you want to be notified about its releases
  * The core is the component that sets the standards and the way things are functioning when it comes to listening
    * It is offering two ways of checking if a new release has been shipped:
      * synchronously - gets the information and just run the checks and returns the information
      * async - gets information about the target and returns a channel that the user can subscribe to 
* Logging standards
* Tracing with OTEL (https://www.jaegertracing.io/)
  * https://medium.com/jaegertracing/experiment-migrating-opentracing-based-application-in-go-to-use-the-opentelemetry-sdk-29b09fe2fbc4
* A basic layer for dependency injection and app state management 

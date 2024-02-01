## Utils

This is a collection of common helper functions we use in building http services in Golang. It provides several services as follows:
- Basic project structure such as `config` object. 
- Also a quick setup for local integration test frameworks with docker & database (postgres).
- Some basic HTTP frameworks regarding authentication & marshaling payload from requests

These are commonly used functions in multiple of the http services we have in the team.
Underlying HTTP framework is [echo](https://github.com/labstack/echo).

## Assumption
This repo assume the project structure as follows:
```
<app_root>
  - config.yml
  - migrations/postgres/<postgres_migrations>
  - <other app code>
```
### App root
The <app root> should be a unique name in the path, and it's relative position with regard to `config.yml` and `migrations/` should be fixed as above.

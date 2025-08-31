# demo-shop

A simple Go based demo microservice application.

## Requirements

- Podman >= 5.5.1

## Quick Start

### Start stack

`make run`

or

`make rund`

to run the stack in detached mode

## Flow Diagram

```mermaid
graph TD;
  A[Frontend] --> B[Gateway];
  B --> C[User Service];
  B --> D[Item Service];
  B --> E[Cart Service];
  B --> F[Cart Presentation];
  B --> G[Checkout Service];

  F --> D;
  F --> E;

  G --> D;
  G --> E;
```

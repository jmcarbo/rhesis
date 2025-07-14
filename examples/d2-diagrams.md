# D2 Diagrams in Rhesis

## Simple Flow

Here's a simple flow diagram:

```d2
start: Start Process
analyze: Analyze Data
decide: Make Decision
execute: Execute Action
end: Complete

start -> analyze
analyze -> decide
decide -> execute
execute -> end
```

Let's explore how D2 diagrams work in presentations.

---
This simple flow shows the basic process from start to finish.

Duration: 10

## System Architecture

```d2
users: Users {
  shape: person
}

lb: Load Balancer {
  shape: hexagon
}

web1: Web Server 1
web2: Web Server 2
web3: Web Server 3

cache: Redis Cache {
  shape: cylinder
}

db: PostgreSQL {
  shape: cylinder
}

users -> lb: HTTPS
lb -> web1: HTTP
lb -> web2: HTTP
lb -> web3: HTTP

web1 -> cache: Read/Write
web2 -> cache: Read/Write
web3 -> cache: Read/Write

web1 -> db: Query
web2 -> db: Query
web3 -> db: Query
```

A typical web application architecture with load balancing.

---
This architecture diagram shows how users connect through a load balancer to multiple web servers, which interact with both a cache layer and a database.

Duration: 15

## State Machine

```d2
idle: Idle {
  shape: circle
}

processing: Processing {
  shape: rectangle
}

error: Error {
  shape: diamond
  style.fill: "#ff6b6b"
}

complete: Complete {
  shape: circle
  style.fill: "#51cf66"
}

idle -> processing: Start Task
processing -> complete: Success
processing -> error: Failure
error -> idle: Reset
complete -> idle: Reset
```

State machines are great for showing application flow.

---
This state machine demonstrates how a system transitions between different states based on various events and conditions.

Duration: 12

## Network Topology

```d2
internet: Internet {
  shape: cloud
}

firewall: Firewall {
  shape: hexagon
  style.fill: "#ffa94d"
}

dmz: DMZ Network {
  web_server: Web Server
  mail_server: Mail Server
}

internal: Internal Network {
  app_server: Application Server
  db_server: Database Server
  file_server: File Server
}

internet -> firewall: Public Traffic
firewall -> dmz: Filtered Traffic
dmz -> internal: Secure Connection
```

Network diagrams help visualize infrastructure.

---
This topology shows a typical network setup with DMZ for public-facing services and an internal network for secure resources.

Duration: 15

## Data Flow Pipeline

```d2
source: Data Sources {
  api: REST API
  files: File System
  stream: Message Queue
}

ingestion: Data Ingestion {
  validate: Validation
  transform: Transformation
}

storage: Data Lake {
  shape: cylinder
  style.multiple: true
}

processing: Processing {
  analytics: Analytics Engine
  ml: ML Pipeline
}

output: Output {
  dashboard: Dashboard
  reports: Reports
  api_out: API
}

source.api -> ingestion.validate
source.files -> ingestion.validate
source.stream -> ingestion.validate

ingestion.validate -> ingestion.transform
ingestion.transform -> storage

storage -> processing.analytics
storage -> processing.ml

processing.analytics -> output.dashboard
processing.analytics -> output.reports
processing.ml -> output.api_out
```

Modern data pipelines involve multiple stages of processing.

---
This diagram illustrates a complete data pipeline from ingestion through processing to final output, showing how data flows through various transformation stages.

Duration: 20

## Conclusion

D2 diagrams provide a powerful way to visualize:
- System architectures
- Data flows
- State machines
- Network topologies
- Any connected relationships

The declarative syntax makes it easy to create and maintain diagrams as code.

---
D2 integration in Rhesis allows you to create dynamic, professional diagrams directly in your presentations using simple text-based syntax.

Duration: 10
# Smol

Just a Smol Go Web Service

## Backend
- SQLite3
- router: chi
- sdk: oto
- authn
    - login: goth
        - [x] Google
    - session: scs
- authz: simple rbac
- proxy: Caddy

## Web App
- ReactJS Vite MPA
- Chakra UI

## System Overview
```mermaid
graph TB
    Client--"Usess Web App"-->Proxy
    subgraph "Smol System"
        Proxy["Proxy"]
        DataStore["Data Store<br> (SQLite3) <br><br> Main Database"]
        AuthService["Auth Service<br>(Go)<br><br>Session Management & Authorization"]
        TodoService["Todo Service<br>(Go) <br><br>Manage Todos"]
        LiteStream["Lite Stream<br>(Go) <br><br>Backup SQLite3 to Object Storage"]
        
        subgraph "Web App System"
            WebApp["Web App<br>(ReactJS)<br><br>End User App"]
            WebServer["Web Server<br> Serve The Web App"]
            
            WebServer-->WebApp
        end
        
        
        Proxy -- "[TCP]" --> WebServer
        
        WebApp -- "[Proxy]" --> TodoService 
        WebApp -- "[Proxy]" --> AuthService 
        
        AuthService -- "[IPC]" --> DataStore
        TodoService -- "[IPC]" --> DataStore -- Backup --> LiteStream
    end
    AWSS3["AWS S3 <br><br>Object Storage"]
    LiteStream -- Store --> AWSS3
```

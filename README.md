# Smol

Just a Smol Go Web Service

## Backend
- postgres with sqlc-pgx
- router: chi
- authn
    - login: goth
        - [x] Google
    - session: scs
- authz: simple rbac
- proxy: Caddy

## Framework Architecture

### System Overview
```mermaid
graph TB
    subgraph "Smol System"
        DataStore["Data Store<br> (PostgreSQL) <br><br> Main Database"]
        AuthService["Auth Service<br>(Go)<br><br>Session Management & Authorization"]
        TodoService["Todo Service<br>(Go) <br><br>Manage Todos"]
        
        Client -- "Manage Todo <br>[HTTP]" --> TodoService 
        Client -- "Login/Register <br>[HTTP]" --> AuthService 
        
        AuthService -- "[IPC]" --> DataStore
        TodoService -- "[IPC]" --> DataStore
    end
```
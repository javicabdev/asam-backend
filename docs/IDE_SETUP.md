# IDE Configuration Guide

Configuration guide for setting up the development environment for ASAM Backend.

## IntelliJ IDEA / GoLand Setup

### SQL Dialect Configuration

The project uses **PostgreSQL** as the database. SQL files should be recognized as PostgreSQL dialect.

#### Automatic Configuration ✅
The project includes:
- `.idea/sqldialects.xml` - Maps `migrations/` directory to PostgreSQL
- SQL files include `--!postgresql` directive for dialect detection
- `.editorconfig` for consistent code formatting

#### Manual Configuration (if needed)
1. Go to **File → Settings → Languages & Frameworks → SQL Dialects**
2. Set **Project SQL Dialect** to **PostgreSQL**
3. Map `migrations/` directory to **PostgreSQL**

### Database Connection (Optional)
1. **Database → New → Data Source → PostgreSQL**
2. Configure connection to local development database
3. Test connection and download drivers if needed

### Recommended Plugins
- **Database Navigator** - Enhanced database tooling
- **Go** - Go language support (should be included)
- **Docker** - For container management
- **ENV File Support** - For .env files

## VS Code Setup (Alternative)

### Extensions
- **PostgreSQL** by Chris Kolkman
- **SQL Tools** by Matheus Teixeira  
- **Go** by Google
- **Docker** by Microsoft

### Settings
```json
{
  "sql.connections": [
    {
      "name": "ASAM Local",
      "driver": "PostgreSQL",
      "server": "localhost",
      "port": 5432,
      "database": "asam_db",
      "username": "your_user"
    }
  ],
  "[sql]": {
    "editor.defaultFormatter": "bradymholt.pgformatter"
  }
}
```

## Database Tools

### Recommended GUI Tools
- **pgAdmin** - Full-featured PostgreSQL administration
- **DBeaver** - Universal database tool
- **Beekeeper Studio** - Modern, lightweight SQL editor
- **TablePlus** - Native database client (macOS/Windows)

### Command Line Tools
- **psql** - PostgreSQL interactive terminal
- **pg_dump** - Database backup utility
- **migrate** - Migration tool (already included in project)

## Project Configuration Files

- **`.idea/sqldialects.xml`** - IntelliJ SQL dialect mapping
- **`.editorconfig`** - Cross-editor configuration
- **`gqlgen.yml`** - GraphQL code generation
- **`.golangci.yml`** - Go linting configuration
- **`.gosec.json`** - Security linting configuration

## Troubleshooting

### SQL Warnings
If you see warnings like "SQL dialect is not configured":
1. Check that `.idea/sqldialects.xml` exists
2. Verify SQL files start with `--!postgresql`
3. Restart IDE if configuration was just added

### Go Module Issues
```bash
go mod tidy
go mod download
```

### Database Connection Issues
1. Ensure Docker is running: `docker-compose up -d`
2. Check environment variables in `.env.development`
3. Verify database exists: `psql -h localhost -U your_user -l`

### Migration Issues
```bash
# Check migration status
go run cmptemp/migrate/main.go -cmd version

# Reset database (development only!)
./scripts/dev/fresh-database-setup.sh
```

## Development Workflow

1. **Start Database**: `docker-compose up -d`
2. **Run Migrations**: `go run cmptemp/migrate/main.go -cmd up`
3. **Seed Data**: `go run cmptemp/seed/main.go`
4. **Start API**: `make run` or `go run cmd/api/main.go`
5. **GraphQL Playground**: http://localhost:8080/graphql

---

> 💡 **Tip**: Most configuration is automated through project files. If you encounter issues, try restarting your IDE first.

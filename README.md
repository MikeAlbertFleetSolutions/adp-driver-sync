# adp-driver-sync

This application synchronizes driver addresses from ADP Workforce Now to the [Mike Albert Fleet Solutions API](https://developer.mikealbert.com/).

## Prerequisites

- **Go 1.25.0+** installed ([download](https://go.dev/dl/))
- ADP API credentials and SSL certificate
- Mike Albert API credentials

## ADP Setup

You will need to have an ADP account with access to the Workforce Now APIs.

### Required ADP Credentials
1. **Client ID** - OAuth2 client ID from ADP Developer Portal
2. **Client Secret** - OAuth2 client secret from ADP Developer Portal
3. **SSL Certificate** (`.crt` file) - Client certificate for mTLS authentication
4. **Private Key** (`.pem` or `.key` file) - Private key matching the certificate

> ⚠️ **Important**: ADP requires mutual TLS (mTLS) authentication. You must have both the certificate and private key files. These are typically provided when you register your application with ADP.

### Required ADP Permissions
- Access to Workforce Now APIs (`/hr/v2/workers`)
- OAuth2 client credentials for API authentication
- Permission to read employee information and addresses

### Data Mapping
The application extracts the following information from ADP Workforce Now:
- Employee Number (from `workerID.idValue`)
- First Name and Last Name (from `person.legalName`)
- Home Address (from `person.legalAddress`)

## Configuration

Create a configuration file named `adp-driver-sync.yaml` with the following structure:

```yaml
adp:
  clientid: "your-adp-client-id"
  clientsecret: "your-adp-client-secret"
  baseurl: "https://api.adp.com"
  certfile: "path/to/your/certificate.crt"
  keyfile: "path/to/your/privatekey.pem"

mikealbert:
  clientid: "your-mike-albert-client-id"
  clientsecret: "your-mike-albert-client-secret"
  endpoint: "https://your-mikealbert-endpoint.com/api/v1"
```

### Configuration Details

| Field | Description |
|-------|-------------|
| `adp.clientid` | OAuth2 Client ID from ADP Developer Portal |
| `adp.clientsecret` | OAuth2 Client Secret from ADP Developer Portal |
| `adp.baseurl` | ADP API base URL (typically `https://api.adp.com`) |
| `adp.certfile` | Path to your ADP SSL certificate file (`.crt`) |
| `adp.keyfile` | Path to your private key file (`.pem` or `.key`) |
| `mikealbert.clientid` | Client ID provided by Mike Albert |
| `mikealbert.clientsecret` | Client Secret provided by Mike Albert |
| `mikealbert.endpoint` | Mike Albert API endpoint URL |

## Running Locally

### 1. Clone the repository
```bash
git clone https://github.com/MikeAlbertFleetSolutions/adp-driver-sync.git
cd adp-driver-sync
```

### 2. Place your certificate files
Copy your ADP certificate and private key files to the project directory (or update the paths in your config):
```
adp-driver-sync/
├── adpCert.crt      # Your ADP certificate
├── adpKey.pem       # Your private key
└── adp-driver-sync.yaml
```

### 3. Configure the application
Edit `adp-driver-sync.yaml` with your credentials (see Configuration section above).

### 4. Build the application
```bash
go build -o adp-driver-sync.exe ./cmd/adp-driver-sync
```

Or on Linux/Mac:
```bash
go build -o adp-driver-sync ./cmd/adp-driver-sync
```

### 5. Run the application
```bash
./adp-driver-sync -config adp-driver-sync.yaml
```

On Windows:
```powershell
.\adp-driver-sync.exe -config adp-driver-sync.yaml
```

## Running as a Scheduled Task

This application can be run as a cron job (Linux/Mac) or scheduled task (Windows) to periodically sync driver information.

### Linux/Mac (cron)
```bash
# Run daily at 2 AM
0 2 * * * /path/to/adp-driver-sync -config /path/to/adp-driver-sync.yaml
```

### Windows (Task Scheduler)
Create a scheduled task that runs:
```
adp-driver-sync.exe -config adp-driver-sync.yaml
```

## API Details

### ADP Workforce Now API
- **Authentication**: OAuth2 Client Credentials Grant with mTLS
- **Token Endpoint**: `POST /auth/oauth/v2/token`
- **Workers Endpoint**: `GET /hr/v2/workers`
- **Data Format**: JSON

### Mike Albert API
- **Authentication**: Client credentials (client_id + client_secret)
- **Token Endpoint**: `POST {endpoint}/token`
- **Find Drivers**: `POST {endpoint}/driver-management/driver/find`
- **Update Driver**: `POST {endpoint}/driver-management/driver/{id}`

## Troubleshooting

### "proper client ssl certificate was not presented"
You need both the certificate (`.crt`) AND the private key (`.pem`) files. Make sure:
- Both files exist at the paths specified in your config
- The private key matches the certificate (they must be from the same key pair)

### "private key does not match public key"
The certificate and private key files are not from the same pair. Contact your ADP administrator to get the matching files.

### "Invalid client_id or client_secret" (Mike Albert)
- Verify your Mike Albert credentials are correct
- Check that the endpoint URL is correct for your environment
- Ensure the config uses `clientid` and `clientsecret` (no underscores)

### DNS resolution errors
If you see "no such host" errors, you may need to be connected to a VPN or corporate network to reach internal endpoints.

## Development

### Build for Linux (from Windows)
```bash
set GOOS=linux
set GOARCH=amd64
go build -o target/adp-driver-sync ./cmd/adp-driver-sync
```

### Run code checks
```bash
go fmt ./...
go vet ./...
```

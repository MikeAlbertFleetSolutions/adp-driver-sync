# adp-driver-sync

This sample application synchronizes your drivers addresses from ADP Workforce Now to the [Mike Albert Fleet Solutions API](https://developer.mikealbert.com/).


## ADP Setup

You will need to have an ADP account with access to the Workforce Now APIs. The application uses ADP's OAuth2 authentication and queries employee data directly from the Workforce Now REST APIs.

### Required ADP Permissions
- Access to Workforce Now APIs (`/hcm/v1/workers`)
- OAuth2 client credentials for API authentication
- Permission to read employee work assignments and address information

### Data Mapping
The application extracts the following information from ADP Workforce Now:
- Employee Number (from ADP Worker ID)
- First Name and Last Name (from person.legalName)
- Home Address (from workAssignments[].homeWorkLocation.address)

## Configuration

The configuration file should be in YAML format and include the following information:

adp:
  clientid: "your-adp-client-id"
  clientsecret: "your-adp-client-secret"
  baseurl: "https://api.adp.com"
mikealbert:
  clientid: "your-mike-albert-client-id"
  clientsecret: "your-mike-albert-client-secret"
  endpoint: "https://api.mikealbert.com/api/v1/"### Configuration Details
- **ADP clientid/clientsecret**: Obtained from ADP Developer Portal for OAuth2 authentication
- **ADP baseurl**: Your ADP API environment URL (typically `https://api.adp.com`)
- **Mike Albert credentials**: Provided by Mike Albert for their API access

## Running the Application

The application can be run from the command line with the following command:

adp-driver-sync -config adp-driver-sync.yaml
This could also be run as a cron job or a scheduled task.

## API Details

The application uses ADP Workforce Now APIs to retrieve current employee information:
- **Authentication**: OAuth2 Client Credentials Grant
- **Endpoint**: `GET /hcm/v1/workers?address=true`
- **Data Format**: JSON responses parsed for address information
- **No custom reports needed**: Uses live ADP employee data directly
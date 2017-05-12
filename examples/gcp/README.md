# GCE Sensu Deployment Example

This example shows how to deploy Sensu to Google Container Enginer and set up
basic monitoring for a simple service.

You will:

1. Build Sensu
2. Deploy Sensu to an existing Google Container Enginer Kubernetes cluster.
3. Deploy an application to the container cluster.
4. Create a Slack handler for alert notifications.
5. Create a check for your new service.
6. Simulate a service failure and receive a notification via Slack.

## Walk-through

### Building Sensu

From the project root, run:

- `./build.sh`
- `./build.sh docker`

### Deploying Sensu

If you have not configured the Google Cloud SDK or launched a GCE cluster,
see the quickstart guide from the GCP team.

https://cloud.google.com/container-engine/docs/quickstart

1. Create a persistent disk for sensu-backend.

`gcloud compute disks create --size 200GB sensu-backend-disk`

2. 

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

### Build Sensu

This is currently specific to our internal monitorama demo GCE project, but can
be easily made generic later.

NOTE: This takes some time. I would make sure this is done _before_ the demo.

From the project root, run:

- `./build.sh`
- `./build.sh docker`
- `docker tag sensu/sensu:latest us.gcr.io/monitorama-166316/sensu`
- `docker build -f Dockerfile.agent -t us.gcr.io/monitorama-166316/sensu-agent .`
- `gcloud docker -- push us.gcr.io/monitorama-166316/sensu`
- `gcloud docker -- push us.gcr.io/monitorama-166316/sensu-agent`

### Deploy Sensu

If you have not configured the Google Cloud SDK or launched a GCE cluster,
see the quickstart guide from the GCP team.

https://cloud.google.com/container-engine/docs/quickstart

From the examples/gce directory:

1. Create a persistent disk for sensu-backend.

`gcloud compute disks create --size 200GB sensu-backend-disk`

2. Launch sensu-backend

`kubectl create -f sensu-backend-deployment.yaml`

3. Expose the sensu-backend externally as a service.

`kubectl create -f sensu-backend-service.yaml`

It can take 30-60 seconds for a public IP address to be assigned to the service's
load balancer.

### Launch a service in GCE

For this example, we're using a special container that has everything Sensu needs
to check the health of the service.

1. Launch the dummy service.

`kubectl create -f dummy-deployment.yaml`

2. Expose the dummy service externally.

`kubectl create -f dummy-service.yaml`

### Create a Handler

NOTE: This is also specific to the example.

In order to create the Slack handler, we need to configure the sensu-cli to talk to
our Sensu service.

1. Get the external IP address of the Sensu backend.

`kubectl get service/sensu-backend`

If the service does not have an IP address yet, you may see output like this:

```
λ kubectl get service/sensu-backend
NAME            CLUSTER-IP     EXTERNAL-IP   PORT(S)                                        AGE
sensu-backend   10.3.247.174   <pending>     8080:30331/TCP,8081:31900/TCP,3000:32374/TCP   4s
```

Keep querying it until you get an external IP assigned. This can take 30-60 seconds. If you're
following the howto script, it should be just about ready by now.

```
λ kubectl get service/sensu-backend
NAME            CLUSTER-IP     EXTERNAL-IP      PORT(S)                                        AGE
sensu-backend   10.3.247.174   104.197.192.46   8080:30331/TCP,8081:31900/TCP,3000:32374/TCP   1m
```

2. Configure sensu-cli

Make sure that sensu-cli is in your path by coping it from the `bin/` directory in the project
root to, for example, /usr/local/bin

`cp ../../bin/sensu-cli /usr/local/bin`

`sensu-cli configure`

```
λ sensu-cli configure
? Sensu Base URL: http://104.197.192.46:8080
? Email: greg@notanemail.poop
? Password:  ******
? Preferred output: yaml
```

3. Create the slack handler

`sensu-cli handler create`

```
λ sensu-cli handler create
? Handler Name: slack
? Mutator:
? Timeout: 60
? Type: pipe
? Command: /opt/bin/handlers/slack -w https://hooks.slack.com/services/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 -c '#sensu-spam'
OK
```

### Create a Check

Now we'll create a check for our dummy service.

1. Create a check

`sensu-cli check create`

```
λ sensu-cli check create
? Check Name: dummy-service
? Command: /opt/sensu/checks/check.sh http://localhost/healthz
? Interval: 10
? Subscriptions: dummy
? Handlers: slack
OK
```


# GCE Sensu Deployment Example

This example shows how to deploy Sensu to Google Container Enginer and set up
basic monitoring for a simple service.

Prerequisite: build Sensu.

You will:

1. Launch an application in Kubernetes.
2. Deploy Sensu to an existing Google Container Enginer Kubernetes cluster.
3. Add Sensu to our existing application deployment.
4. Scale your application.
5. Create a Slack handler for alert notifications.
6. Create a check for your new service.
7. Verify the check is running.
8. Simulate a failure.

## Prerequisites 

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

### GCP/GCE Setup

If you have not configured the Google Cloud SDK or launched a GCE cluster,
see the quickstart guide from the GCP team.

https://cloud.google.com/container-engine/docs/quickstart

## Walk-through

### Launch an App

For this example, we're using a special container that has everything Sensu needs
to check the health of the service.

See: http://github.com/grepory/dummy

1. Launch the dummy service.

`kubectl create -f dummy-deployment.yaml`

NOTES: 
- show the dummy-deployment.yaml contents.
- explain how a deployment works:
  - a deployment creates a replica set, we've specified we only want one replica
  - the pod specification comes next, show the container name, image name, explain
    the exposed ports. (it listens on port 80 and that port gets exposed).

2. Expose the dummy service externally.

`kubectl create -f dummy-service.yaml`

NOTES:
- The service type is a LoadBalancer -- The GCE integration with Kubernetes provides a
  GCP load balancer with an external IP address.
- The ports specified in the YAML are commented so that you know which port does what.

### Deploy Sensu

From the examples/gce directory:

1. Create a persistent disk for sensu-backend.

`gcloud compute disks create --size 200GB sensu-backend-disk`

Creating a persistent disk doesn't format the disk right away, so there is no
filesystem. However, GCE, upon first mounting the disk, takes care of this
for us, so there's no need to format the disk ourselves.

2. Launch sensu-backend

`kubectl create -f sensu-backend-deployment.yaml`

NOTES:
- Walk through the Sensu startup process. Explain that only one process
  is responsible for Etcd, the API, the Transport, and the Dashboard!
  (You can see this is the case in the `command` field in the pod spec
  in the deployment YAML.)
- May be worth highlighting that the GCE GCP integration provides us
  with persistent storage so that if our sensu-backend pods move around
  the GCE cluster, they don't lose state.

3. Expose the sensu-backend externally as a service.

`kubectl create -f sensu-backend-service.yaml`

NOTES:
- It can take 30-60 seconds for a public IP address to be assigned to the service's
load balancer. That needs to complete before you continue.

4. Get the external IP address of the Sensu backend.

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

5. Configure sensu-cli

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

### Add Sensu to your App

Now we want to start monitoring our application. So we're just going to add a
sensu-agent sidecar to our app's pod. For our example, the agent's container
has a check baked into the container, it's very simple, you can even show it.

Paths relative to project root.j
Check source: examples/checks/http\_check.sh
Agent container dockerfile: Dockerfile.agent

1. Replace our deployment specification with the new one including Sensu.

`kubectl replace -f dummy-with-sensu.yaml`

NOTES:
- Walk through the new pod specification.

2. Verify that the sensu-agent process is running and registered with 
   the backend.

NOTE: This isn't in the CLI yet, but you can showcase that sensu-backend
is API driven using curl instead. I piped it through JQ to do syntax
highlighting to make it more readable.

`curl http://104.197.192.46:8080/entities | jq`

### Create a Handler

1. Create the slack handler

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

NOTES:
- Walk through the check, it's in examples/checks/http\_check.sh
  It's just a basic curl that returns a non-zero exit code if the GET
  returns a non-200 status.

### Verify the Check is Running

The check should be running within seconds. You can then verify that the
check is running and see the state of your service pods:

```
λ sensu-cli event list
Source               Check    Result               Timestamp
──────────────────────────────── ─────── ─────────── ───────────────────────────────
dummy-service-2833516106-6xwq2   dummy   healthy     2017-05-12 13:34:05 -0700 PDT
```


### Simulate an Event

The dummy service accepts POST HTTP requests to its /healthz endpoint which will
toggle the health of the service. We're going to simulate a failure of our service
so that we can see the handler emit an event to the #sensu-spam Slack channel.

We'll do this by curling from within the pod's sensu-agent container which has
curl installed already.

```
λ kubectl exec -it -c sensu-agent dummy-service-2833516106-dg776 /bin/sh
/ # curl -X POST http://localhost/healthz
/ # curl -X GET http://localhost/healthz
unhealthy
```

You should see a Slack alert for the check pretty quickly. Because the check
interval is 10 seconds, though, you're going to get a Slack alert every 10
seconds.

NOTES:
- You can explain how the dummy service works. POSTing to the /healthz endpoint
  will cause the service to say that it's failing.
- If you leave the kubectl exec open, you can POST again to make the check pass
  again, but when I tried this, I didn't see a resolution notification. I don't
  know if that's a bug or okay.


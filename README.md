# k8s-go

A collection of ad hoc scripts for kubernetes using the go client.

CD in each directory folder under `/scrips` and run `go run main.go`

All scripts will authenticate with to cluster with local kube config.

- replicas

List deployments with less than 2 replicas for quick high availability check.

- resources

List all pods and jobs that not have any CPU/MEM request and limits set.

- vpa-deployment

Iterate over deployments in cluster and create a Virtual pod autoscaling resource (VPA) in update mode off.

- vpa-ds

Iterate over daemonsets in cluster and create a Virtual pod autoscaling resource (VPA) in update mode off.

- vpa-sts

Iterate over daemonsets in cluster and create a Virtual pod autoscaling resource (VPA) in update mode off.
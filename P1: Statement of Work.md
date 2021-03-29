Part 1: Execute Build Script in the Cloud
---------------------------------

# Purpose

The build system of Kerbodyne's Mastodon engine microcontroller has grown beyond the reasonable compute capacity of engineer workstations. The monorepo containing all the code to generate the controller binary are stored in `github.com/kerbodyne/ke-1`. 

We need these builds to be run in the cloud on our Kubernetes cluster, the task is to create Makefile targets which will automate the process of building, waiting for completion, and enabling user access to artifacts.

[Kubernetes](https://kubernetes.io/docs/home/)(k8s) is the compute resource manager we'll be using to run these builds via the [Jobs](https://kubernetes.io/docs/concepts/workloads/controllers/job/) tooling. Kubernetes Jobs are run till completion, reguardless of Node failures, so even if the cluster is having issues we should always get a result from our build. Kubernetes is a massive ecosystem of tools and evolving features, so explore functionality and options; recommend changes if you find improvements!

# Tasks defined as [Makefile](https://www.gnu.org/software/make/manual/make.html) targets

### Build Identification

All steps of the build pipeline should be tagged with the source control `git rev-parse --short HEAD` hash or tag. eg: `c39715e1` or `v1.0.2`: hearby referred to as `BUILD_TAG` This will make tracking errors and artifacts consistent.

* Docker image
* Azure Blob Store Directory

### `make build` - existing target

Existing functionality. Executes build on the source code and models within this repo producing a binary artifact; `ke-1/bin/ke-1-ctl`. If there is a build failure, this target returns error output.

## `make build-image` - create a Docker Image

Using a [Dockerfile](https://docs.docker.com/develop/develop-images/dockerfile_best-practices/), wrap up necessary source code, Makefile, packages and libraries to run `make build`. Creating an image capable of running `make build-wrapper`, and push image to [Github Container Registry](https://docs.github.com/en/packages/guides/about-github-container-registry).

* Docker image is tagged with BUILD_TAG, eg: `docker.pkg.github.com/kerbodyne/ke-1-build:BUILD_TAG`
* Image run entrypoint is `make build-wrapper`.
* Image is pushed to Github's Container Repository(could be done via CI job).

## `make build-wrapper` - run `make build`; archive artifacts

A Makefile target which invokes `make build`, and then exports binaries to Azure Blob Store for access upon completion, using a Python 3 script to interface with Azure via the [Python SDK](https://github.com/Azure/azure-sdk-for-python).

In the case of failure or success, logs from `make build`, should be picked up by the Python script and written to [Azure blob store](https://azure.microsoft.com/en-us/services/storage/blobs/) account for access by developers, users, and CI/CD systems.

* eg: `make build 2>&1 | tee /tmp/log-$(date +%F:+%T)` 
    * **note**: `tee` will silence an error code returned from the make command, this works to our advantage inside a Kubernetes Job, allowing the next Make target to run.

* Format: `http://kerbodyne.blob.core.windows.net/ke-1-builds/BUILD_TAG/` with subdirectories: 
    * `/bin/` - Binary artifacts
    * `/logs/` - Log files named by Timestamp

In order to Authenticate with Azure Blob API from the Container, the process will need to be provided with a [Pod-managed Identity](https://docs.microsoft.com/en-us/azure/aks/operator-best-practices-identity#use-pod-managed-identities). Ask the Security team for a cluster available Pod Identity with Write Access to the Blob store Container, and [apply it](https://docs.microsoft.com/en-us/azure/aks/use-azure-ad-pod-identity#run-a-sample-application) to the [Kubernetes Job Spec](https://kubernetes.io/docs/concepts/workloads/controllers/job/).

## `make launch-build`

The command to launch all the components together and wait for completion!

Using [Kubernetes Kustomize](https://kubernetes.io/docs/tasks/manage-kubernetes-objects/kustomization/), create a [Job Spec](https://kubernetes.io/docs/concepts/workloads/controllers/job/) template, which will be configured to run the previously built Docker image:BUILD_TAG. The Job spec will also need to include the Pod Identity mentioned above to grant security access to Blob store. 

Kustomize is a bit of a large project, but it is a nice way to extend and template Kubernetes declarations. Write Kustomize's output to a file, eg: `job-spec.yaml` 

Kustomize will output a YAML declaration, which can then be created in the cluster with `kubectl create -f job-spec.yaml`. Once the Job is created, Kubernetes will pull the image and execute the `make build-wrapper` target in the container, and signal completed once it's run to completion.

In this Makefile target, a user can wait for the Job's completion with:

* `kubectl wait --for=condition=complete --timeout=6000s kerbodyne/ke-1-build`
    * This command releases once the Job has completed, then the artifacts should be available in the Azure Blob Store.

* Artifacts
    * `ke-1-ctl` binary
    * Logs - named with timestamp in Blob store.

Final step; collect the artifacts using from the Blob store Container `/ke-1-builds/BUILD_TAG/...`.

This could be achieved with the `azcopy` [CLI tool](https://docs.microsoft.com/en-us/azure/storage/common/storage-ref-azcopy-copy), pulling the files down into the user's local machine. eg: `azcopy cp "https://[account].blob.core.windows.net/[container]/[path/to/blob]" "/path/to/file.txt"` This could be separated into a separate Make target.


# Deliverable

A deterministic toolchain to deploy our source code build into a Kubernetes cluster, and collect the artifacts for the user.

I've outlined the tasks as a set of Make targets, however it might be more efficient to break the steps down further, and apply Python scripts to execute all of the Azure integrations. 

### Priorities

* **Security**
    * This project deals with our organization assets. Take every precaution to keep assets used and produced secured within our Authenticated systems, and limit Authorization tokens by least privilege necessary to complete the task.
* **Determinism**
    * This is a build system, being able to reproduce errors quickly and efficiently will save work hours and mental stress. Limit or eliminate the number of assets loaded into the container at runtime, and instead save them into the Docker Image during the build process.
    * Use the git commit hash of the build's source code to tag and label all images and artifacts.
* **Learn**
    * I haven't got into high depth of Kubernetes, so explore the documentation, and maybe you'll find improvements to make!

### Optimization Ideas

* {Script} pre-flight check to Blob store to assert the commit hash has not already been built.
* Cleanup of Job container artifacts via [ttl configuration](https://kubernetes.io/docs/concepts/workloads/controllers/job/#ttl-mechanism-for-finished-jobs).
* Notify monitoring system Prometheus every time a build starts, then the completions and failures.
* Integrate with our CI system.

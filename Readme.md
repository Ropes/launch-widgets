Launch-Widgets
--------------

### Part 1 

Written to `P1: Statement of Work.md`. 

### Part 2

Written in Go, pull down dependencies with `make mod` can be invoked by `make run`, tests by `make test`, and source code can be found in pkg/schedule, a main.go entrypoint with example runtime code can be found in `cmd/`

## Running the code

Make targets are provided to source dependencies, execute tests, and run an example program. 

* `make mod` will ensure the few dependencies are available to compile.
* `make test` runs unit tests.
* `make run` executes the small program in `cmd/main.go`. It is configured with a few hardcoded parameters to show the Pod ComputeResource constraints interacting balancing with the Pod Priority scores on a over-utilized cluster of Nodes. Pods will periodically be scheduled and then unscheduled as their priority is lowered by running on Nodes, thus enabling previously unscheduled Pods to replace them.


### Node

Nodes encapsulate 

* Resources 
    * Capacity Limit(Maximum)

* Start(Pod) error{}
* Stop(Pod) error{}
    * evict(application)
* InvoicePod(Pod) { "commit charges, reduce priority of Pod due to time running" }

Count assigned Apps "hosted" by node

Add/Remove App from Node
    * Consume/Release capacity

### Pod/App

Container data object, unit of work to be assigned 

* ID: UUID for tracking
* Resources Requested
    * mCPU
    * Memory bytes
* Priority: Scalar value indicating Priority among other Pods, 0 being hightest priority.
* Labels() map[string]string{ "name": PodN }

## Scheduler

* Ideally use [Dynamic Double-Auction Algorithm](https://www.cs.cmu.edu/~sandholm/cs15-892F15/Chain-DynamicDouble07.pdf)
    * Scheduler in pkg/schedule uses container priority, and Node capacity compared to Pod requirements, to determine whether a pod can be scheduled.

### Pod Priority 
Use `priority` value for simplicity, with 0 being highest priority to be scheduled. All Apps start with a pValue, and over time the nodes increment their pValues, thus decreasing their scheduling priority. 

Scheduler scans all the Pods pValues, and compares which have the lowest and need to be scheduled. This creates an ordered list of Pods to be scheduled, separate from assignment.

### Pod Assignment

Assignment utilizes the Pod ordering previously determined, and iterates the set of Nodes; assigning Pods to be run. Highest priority get assigned first, round robin'ing through the list of Nodes. Low priority Pods, or Pods with compute requirements beyond Node size, will fail to be assigned if there is insufficient capacity.

* Schedule([]Nodes, []Pods)
    * Iterate over Pods/Apps 
        * Iterate over Nodes to assign
            * If Node has capacity, start Pod
            * Else check next nodes until all Nodes have been checked as candidates for this Pod.



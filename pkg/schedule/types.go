package schedule

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type ComputeResource struct {
	mCPU int
	mem  int
	// Storage uint
	// NetworkBandwidth/s uint
}

func addComputeResources(crs []ComputeResource) ComputeResource {
	n := ComputeResource{}
	for _, c := range crs {
		n.mCPU += c.mCPU
		n.mem += c.mem
	}
	return n
}

// capacityAbove determine's the capacity available above the parameter value
func (cr ComputeResource) capacityAbove(c ComputeResource) ComputeResource {
	if cr.mCPU > c.mCPU && cr.mem > c.mem {
		return ComputeResource{
			mCPU: cr.mCPU - c.mCPU,
			mem:  cr.mem - c.mem,
		}
	}
	mCPU := cr.mCPU - c.mCPU
	if mCPU < 0 {
		mCPU = 0
	}
	mem := cr.mem - c.mem
	if mem < 0 {
		mem = 0
	} //could be done with math.abs(), but nearly equivalent LoC due to type conversions
	return ComputeResource{
		mCPU: mCPU,
		mem:  mem,
	}
}

func (cr *ComputeResource) greaterThan(c *ComputeResource) bool {
	if cr.mCPU > c.mCPU && cr.mem > c.mem {
		return true
	}
	return false
}

// Pod uniquely identifies an Application of work, and tracks
// its Priority(0 being highest priority).
type Pod struct {
	ID     uuid.UUID
	Labels map[string]string
	CR     ComputeResource
	// Image(s)
	Priority int
}

func NewPod(labels map[string]string, cr ComputeResource, priority int) *Pod {
	p := &Pod{
		Labels:   labels,
		CR:       cr,
		Priority: priority,
		ID:       uuid.New(),
	}
	return p
}

func (p Pod) String() string {
	return fmt.Sprintf("%v-%v-%d-%#v", p.Labels, p.ID, p.Priority, p.CR)
}

// ByPriority is a type to sort nodes based on their Priority field.
type ByPriority []*Pod

func (a ByPriority) Len() int           { return len(a) }
func (a ByPriority) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByPriority) Less(i, j int) bool { return a[i].Priority < a[j].Priority }

// Node has a set of resources(Limit) which can host Pod workloads.
// Each node can specify a PriorityCost multiplier which will apply
// cost to a Pod when InvoicePods() is called. This priority is
// utilized by the Scheduler to remove/replace Pods with those of
// higher priority.
type Node struct {
	Labels map[string]string
	Limit  ComputeResource
	Pods   map[uuid.UUID]*Pod // []*Pod

	PriorityCost int
	log          *log.Logger
}

func (n *Node) String() string {
	ps := []string{}
	for _, v := range n.Pods {
		ps = append(ps, v.String())
	}
	podStr := strings.Join(ps, "\n")
	return fmt.Sprintf("%v-%d-%#v\n%s", n.Labels, n.PriorityCost, n.Limit, podStr)
}

func (n *Node) calculateComputeCost(p *Pod) int {
	cpuPercent := (float64(p.CR.mCPU) / float64(n.Limit.mCPU)) * 100.0
	memPercent := (float64(p.CR.mem) / float64(n.Limit.mem)) * 100.0
	avg := (cpuPercent + memPercent) / 2.0

	return int(float64(n.PriorityCost) * avg)
}

// Capacity available for use on Node.
func (n *Node) Capacity() ComputeResource {
	cr := []ComputeResource{}
	for _, p := range n.Pods {
		cr = append(cr, p.CR)
	}
	used := addComputeResources(cr)

	return n.Limit.capacityAbove(used)
}

// Start a Pod on the Node.
// Does not calculate compute cost.
func (n *Node) Start(p *Pod) error {
	n.Pods[p.ID] = p
	return nil
}

// Stop a pod and remove it from Node list.
func (n *Node) Stop(p *Pod) error {
	delete(n.Pods, p.ID)
	return nil
}

// InvoicePods increments the compute cost priority value.
func (n *Node) InvoicePods() {
	for _, p := range n.Pods {
		cost := n.calculateComputeCost(p)
		n.log.WithFields(log.Fields{"node": n.Labels["name"], "pod": p, "cost": cost}).Debug("priority cost applied to pod")
		p.Priority += cost
	}
}

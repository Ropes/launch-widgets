package schedule

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestComputeResources(t *testing.T) {
	a := ComputeResource{
		mCPU: 500,
		mem:  800,
	}
	b := ComputeResource{
		mCPU: 1000,
		mem:  1000,
	}
	c := ComputeResource{
		mCPU: 100,
		mem:  100,
	}

	t.Run("assert gt", func(t *testing.T) {
		assert.True(t, b.greaterThan(&c))
		assert.True(t, b.greaterThan(&a))
		assert.False(t, b.greaterThan(&b))
		assert.False(t, c.greaterThan(&b))
	})

	t.Run("assert diff", func(t *testing.T) {
		r := a.capacityAbove(b)
		if r.mCPU > 0 {
			t.Error("mCPU should be zeroed", r.mCPU)
		}
		if r.mem > 0 {
			t.Error("mem should be zeroed")
		}

		r = c.capacityAbove(b)
		assert.Equal(t, r.mCPU, 0)
		assert.Equal(t, r.mem, 0)

		r = b.capacityAbove(b)
		assert.Equal(t, r.mCPU, 0)
		assert.Equal(t, r.mem, 0)

		r = b.capacityAbove(c)
		assert.Equal(t, 900, r.mCPU)
		assert.Equal(t, 900, r.mem)

	})
}

func TestNodeCapacity(t *testing.T) {
	a := ComputeResource{
		mCPU: 500,
		mem:  1000,
	}
	b := ComputeResource{
		mCPU: 1000,
		mem:  1000,
	}
	c := ComputeResource{
		mCPU: 100,
		mem:  100,
	}
	d := ComputeResource{
		mCPU: 3000,
		mem:  3000,
	}

	n := Node{
		Limit:        b,
		Pods:         make(map[uuid.UUID]*Pod, 0),
		PriorityCost: 100,
	}

	pod0 := NewPod(map[string]string{}, a, 0)
	pod1 := NewPod(map[string]string{}, d, 0)
	pod2 := NewPod(map[string]string{}, c, 0)

	t.Run("compute cost", func(t *testing.T) {
		cc := n.calculateComputeCost(pod0)
		assert.Greater(t, cc, 0)

		cc = n.calculateComputeCost(pod1)
		assert.Greater(t, cc, n.PriorityCost)
	})

	t.Run("node capacity", func(t *testing.T) {
		zeroCapacity := n.Capacity()
		err := n.Start(pod2)
		assert.NoError(t, err)
		runningCapacity := n.Capacity()

		gt := zeroCapacity.greaterThan(&runningCapacity)
		assert.True(t, gt)

		t.Run("remove non-existant pod", func(t *testing.T) {
			n.Stop(pod1)
			runningCapacity = n.Capacity()
			gt := zeroCapacity.greaterThan(&runningCapacity)
			assert.True(t, gt)
		})

		t.Run("assert full capacity returned after pod removed", func(t *testing.T) {
			n.Stop(pod2)
			runningCapacity = n.Capacity()
			gt = zeroCapacity.greaterThan(&runningCapacity)
			assert.False(t, gt)
		})
	})
}

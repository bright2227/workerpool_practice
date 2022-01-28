package workerpool

import (
	"time"
)

type PoolManager struct {
	CheckDuration  time.Duration
	minWorkerRatio float64
	decreaseRatio  float64
	increaseRatio  float64
}

type Action int

const (
	Increase Action = iota
	Idle
	Decrease
)

func (pm *PoolManager) Manage(p *Pool) {
	heartbeat := time.NewTicker(pm.CheckDuration)
	defer heartbeat.Stop()

	for range heartbeat.C {
		if p.closing {
			p.close()
			break
		}
		act := pm.GetAction(p)
		num := 0
		switch act {
		case Increase:
			num = pm.GetIncreaseNum(p)
			p.addWorker(num)
		case Decrease:
			num = pm.GetDecreaseNum(p)
			p.releaseWorker(num)
		case Idle:
			continue
		}

	}
}

func (pm *PoolManager) GetAction(p *Pool) Action {
	pcap := p.Capacity
	idlenum := len(p.workers)
	if idlenum < int(pcap*1/10) {
		return Increase
	}
	if idlenum > int(pcap*6/10) {
		return Decrease
	}
	return Idle
}

func (pm *PoolManager) GetDecreaseNum(p *Pool) int {
	idlenum := float64(len(p.workers))
	if idlenum <= float64(p.MaxCapacity)*pm.minWorkerRatio {
		return 0
	}
	return int(idlenum * pm.decreaseRatio)
}

func (pm *PoolManager) GetIncreaseNum(p *Pool) int {
	availnum := float64(p.MaxCapacity - p.Capacity)
	if availnum <= float64(p.MaxCapacity)*pm.minWorkerRatio {
		return int(availnum)
	}
	return int(availnum * pm.increaseRatio)
}

package mlp

// ResumeManager handles session resume.
type ResumeManager struct {
	Parked         bool
	LastAppliedID  uint64
	ChainEpoch     uint32
	PendingCells   uint64
}

func NewResumeManager() *ResumeManager {
	return &ResumeManager{}
}

func (r *ResumeManager) Park(lastAppliedID uint64, epoch uint32, pending uint64) {
	r.Parked = true
	r.LastAppliedID = lastAppliedID
	r.ChainEpoch = epoch
	r.PendingCells = pending
}

func (r *ResumeManager) Resume() (uint64, uint32, uint64, bool) {
	if !r.Parked {
		return 0, 0, 0, false
	}
	r.Parked = false
	return r.LastAppliedID, r.ChainEpoch, r.PendingCells, true
}

func (r *ResumeManager) IsParked() bool { return r.Parked }

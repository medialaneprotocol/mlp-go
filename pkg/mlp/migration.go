package mlp

type MigrationState int

const (
	MigrationIdle MigrationState = iota
	MigrationProbing
	MigrationMigrated
	MigrationFailed
)

type PathMigration struct {
	State         MigrationState
	ActivePath    uint8
	CandidatePath uint8
	Challenge     uint64
	Failures      uint32
}

func NewPathMigration(activePath uint8) *PathMigration {
	return &PathMigration{State: MigrationIdle, ActivePath: activePath}
}

func (m *PathMigration) StartProbe(candidate uint8, challenge uint64) {
	m.State = MigrationProbing
	m.CandidatePath = candidate
	m.Challenge = challenge
}

func (m *PathMigration) Complete(response uint64) bool {
	if m.State != MigrationProbing {
		return false
	}
	if response == m.Challenge {
		m.ActivePath = m.CandidatePath
		m.State = MigrationMigrated
		return true
	}
	m.State = MigrationFailed
	m.Failures++
	return false
}

func (m *PathMigration) IsProbing() bool { return m.State == MigrationProbing }

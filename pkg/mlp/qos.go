package mlp

type QosTier int

const (
	QosBackground QosTier = 0
	QosBestEffort QosTier = 1
	QosInteractive QosTier = 2
	QosRealtime   QosTier = 3
)

func (q QosTier) MlpPriority() uint8 {
	switch q {
	case QosBackground:
		return 7
	case QosBestEffort:
		return 6
	case QosInteractive:
		return 5
	case QosRealtime:
		return 2
	}
	return 6
}

func (q QosTier) Weight() uint32 {
	switch q {
	case QosBackground:
		return 1
	case QosBestEffort:
		return 2
	case QosInteractive:
		return 4
	case QosRealtime:
		return 8
	}
	return 1
}

package materialize

type EgressIpsBuilder struct {
	egressIps string
}

func NewEgressIpsBuilder(egressIps string) *EgressIpsBuilder {
	return &EgressIpsBuilder{
		egressIps: egressIps,
	}
}

func ReadEgressIpsDatasource() string {
	return "SELECT egress_ip FROM materialize.mz_catalog.mz_egress_ips;"
}

package materialize

func ReadEgressIpsDatasource() string {
	return "SELECT egress_ip FROM materialize.mz_catalog.mz_egress_ips;"
}

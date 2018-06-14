package mssql

import "open-service-broker-azure/pkg/service"

func (
	d *dbmsManager,
) getProvisionParametersSchema() service.InputParametersSchema {
	return getDBMSCommonProvisionParamSchema()
}

type dbmsInstanceDetails struct {
	ARMDeploymentName        string `json:"armDeployment"`
	FullyQualifiedDomainName string `json:"fullyQualifiedDomainName"`
	ServerName               string `json:"server"`
	AdministratorLogin       string `json:"administratorLogin"`
}

type secureDBMSInstanceDetails struct {
	AdministratorLoginPassword string `json:"administratorLoginPassword"`
}

package mysql

import (
	"open-service-broker-azure/pkg/service"
)

func (d *databaseManager) Unbind(
	instance service.Instance,
	binding service.Binding,
) error {
	pdt := dbmsInstanceDetails{}
	if err :=
		service.GetStructFromMap(instance.Parent.Details, &pdt); err != nil {
		return err
	}
	spdt := secureDBMSInstanceDetails{}
	if err :=
		service.GetStructFromMap(instance.Parent.SecureDetails, &spdt); err != nil {
		return err
	}
	dt := databaseInstanceDetails{}
	if err := service.GetStructFromMap(instance.Details, &dt); err != nil {
		return err
	}
	bd := bindingDetails{}
	if err := service.GetStructFromMap(binding.Details, &bd); err != nil {
		return err
	}
	return unbind(
		isSSLRequired(*instance.Parent.ProvisioningParameters),
		d.sqlDatabaseDNSSuffix,
		pdt.ServerName,
		spdt.AdministratorLoginPassword,
		pdt.FullyQualifiedDomainName,
		dt.DatabaseName,
		bd,
	)
}

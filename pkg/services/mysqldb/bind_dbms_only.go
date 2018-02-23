package mysqldb

import (
	"fmt"

	"github.com/Azure/open-service-broker-azure/pkg/service"
)

func (d *dbmsOnlyManager) ValidateBindingParameters(
	service.BindingParameters,
	service.SecureBindingParameters,
) error {
	// There are no parameters for binding to MySQL, so there is nothing
	// to validate
	return nil
}

func (d *dbmsOnlyManager) Bind(
	instance service.Instance,
	_ service.BindingParameters,
	_ service.SecureBindingParameters,
) (service.BindingDetails, service.SecureBindingDetails, error) {
	return nil, nil, fmt.Errorf("service is not bindable")
}

func (d *dbmsOnlyManager) GetCredentials(
	_ service.Instance,
	_ service.Binding,
) (service.Credentials, error) {
	return nil, fmt.Errorf("service not bindable")
}
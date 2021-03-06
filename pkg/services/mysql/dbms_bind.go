package mysql

import (
	"fmt"

	"github.com/Azure/open-service-broker-azure/pkg/service"
)

func (d *dbmsManager) Bind(
	instance service.Instance,
	_ service.BindingParameters,
) (service.BindingDetails, service.SecureBindingDetails, error) {
	return nil, nil, fmt.Errorf("service is not bindable")
}

func (d *dbmsManager) GetCredentials(
	service.Instance,
	service.Binding,
) (service.Credentials, error) {
	return nil, fmt.Errorf("service not bindable")
}

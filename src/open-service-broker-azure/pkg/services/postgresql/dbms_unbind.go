package postgresql

import (
	"fmt"

	"open-service-broker-azure/pkg/service"
)

func (a *dbmsManager) Unbind(
	instance service.Instance,
	binding service.Binding,
) error {
	return fmt.Errorf("service is not bindable")
}

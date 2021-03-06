package mysql

import (
	"context"
	"fmt"

	"github.com/Azure/open-service-broker-azure/pkg/generate"
	"github.com/Azure/open-service-broker-azure/pkg/service"
	_ "github.com/go-sql-driver/mysql" // MySQL driver
	uuid "github.com/satori/go.uuid"
)

func (a *allInOneManager) GetProvisioner(
	service.Plan,
) (service.Provisioner, error) {
	return service.NewProvisioner(
		service.NewProvisioningStep("preProvision", a.preProvision),
		service.NewProvisioningStep("deployARMTemplate", a.deployARMTemplate),
	)
}

func (a *allInOneManager) preProvision(
	ctx context.Context,
	_ service.Instance,
) (service.InstanceDetails, service.SecureInstanceDetails, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	serverName, err := getAvailableServerName(
		ctx,
		a.checkNameAvailabilityClient,
	)
	if err != nil {
		return nil, nil, err
	}
	dt := allInOneInstanceDetails{
		dbmsInstanceDetails: dbmsInstanceDetails{
			ARMDeploymentName: uuid.NewV4().String(),
			ServerName:        serverName,
		},
		DatabaseName: generate.NewIdentifier(),
	}

	sdt := secureAllInOneInstanceDetails{
		secureDBMSInstanceDetails: secureDBMSInstanceDetails{
			AdministratorLoginPassword: generate.NewPassword(),
		},
	}

	dtMap, err := service.GetMapFromStruct(dt)
	if err != nil {
		return nil, nil, err
	}
	sdtMap, err := service.GetMapFromStruct(sdt)
	return dtMap, sdtMap, err
}

func (a *allInOneManager) deployARMTemplate(
	_ context.Context,
	instance service.Instance,
) (service.InstanceDetails, service.SecureInstanceDetails, error) {
	dt := allInOneInstanceDetails{}
	if err := service.GetStructFromMap(instance.Details, &dt); err != nil {
		return nil, nil, err
	}
	sdt := secureAllInOneInstanceDetails{}
	if err := service.GetStructFromMap(instance.SecureDetails, &sdt); err != nil {
		return nil, nil, err
	}

	version := instance.Service.GetProperties().Extended["version"].(string)

	goTemplateParameters, err := buildGoTemplateParameters(
		instance.Plan,
		version,
		dt.dbmsInstanceDetails,
		sdt.secureDBMSInstanceDetails,
		*instance.ProvisioningParameters,
	)

	if err != nil {
		return nil, nil, fmt.Errorf(
			"error building go template parameters :%s",
			err,
		)
	}
	goTemplateParameters["databaseName"] = dt.DatabaseName
	tagsObj := instance.ProvisioningParameters.GetObject("tags")
	tags := make(map[string]string, len(tagsObj.Data))
	for k := range tagsObj.Data {
		tags[k] = tagsObj.GetString(k)
	}
	outputs, err := a.armDeployer.Deploy(
		dt.ARMDeploymentName,
		instance.ProvisioningParameters.GetString("resourceGroup"),
		instance.ProvisioningParameters.GetString("location"),
		allInOneARMTemplateBytes,
		goTemplateParameters,
		map[string]interface{}{},
		tags,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("error deploying ARM template: %s", err)
	}

	var ok bool
	dt.FullyQualifiedDomainName, ok = outputs["fullyQualifiedDomainName"].(string)
	if !ok {
		return nil, nil, fmt.Errorf(
			"error retrieving fully qualified domain name from deployment: %s",
			err,
		)
	}

	dtMap, err := service.GetMapFromStruct(dt)
	return dtMap, instance.SecureDetails, err
}

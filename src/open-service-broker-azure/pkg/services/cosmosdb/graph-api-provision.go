// +build experimental

package cosmosdb

import (
	"context"
	"fmt"

	"open-service-broker-azure/pkg/service"
)

func (g *graphAccountManager) GetProvisioner(
	service.Plan,
) (service.Provisioner, error) {
	return service.NewProvisioner(
		service.NewProvisioningStep(
			"preProvision", g.preProvision),
		service.NewProvisioningStep("deployARMTemplate", g.deployARMTemplate),
	)
}

func (g *graphAccountManager) deployARMTemplate(
	ctx context.Context,
	instance service.Instance,
) (service.InstanceDetails, service.SecureInstanceDetails, error) {

	pp := &provisioningParameters{}
	if err :=
		service.GetStructFromMap(instance.ProvisioningParameters, pp); err != nil {
		return nil, nil, err
	}

	dt := &cosmosdbInstanceDetails{}
	if err := service.GetStructFromMap(instance.Details, &dt); err != nil {
		return nil, nil, err
	}

	p, err := g.buildGoTemplateParams(instance, "GlobalDocumentDB")
	if err != nil {
		return nil, nil, fmt.Errorf("error building arm params: %s", err)
	}
	p["capability"] = "EnableGremlin"
	if instance.Tags == nil {
		instance.Tags = make(map[string]string)
	}
	instance.Tags["defaultExperience"] = "Graph"
	fqdn, sdt, err := g.cosmosAccountManager.deployARMTemplate(ctx, instance, p)
	if err != nil {
		return nil, nil, fmt.Errorf("error deploying ARM template: %s", err)
	}
	dt.FullyQualifiedDomainName = fqdn
	sdt.ConnectionString = fmt.Sprintf("AccountEndpoint=%s;AccountKey=%s;",
		dt.FullyQualifiedDomainName,
		sdt.PrimaryKey,
	)
	dtMap, err := service.GetMapFromStruct(dt)
	if err != nil {
		return nil, nil, err
	}
	sdtMap, err := service.GetMapFromStruct(sdt)
	return dtMap, sdtMap, err
}
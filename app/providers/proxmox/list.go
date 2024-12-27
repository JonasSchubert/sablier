package proxmox

import (
	"context"
	"fmt"

	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/Telmate/proxmox-api-go/proxmox"

	"github.com/sablierapp/sablier/app/discovery"
	"github.com/sablierapp/sablier/app/providers"
	"github.com/sablierapp/sablier/app/types"
)

func (provider *ProxmoxProvider) InstanceList(ctx context.Context, options providers.InstanceListOptions) ([]types.Instance, error) {
	guests, err := proxmox.ListGuests(cli.Context(), &provider.Client)
	if err != nil {
		return nil, err
	}

	labels := []string{}
	for _, label := range options.Labels {
		labels = append(labels, label)
		labels = append(labels, fmt.Sprintf("%s=true", label))
	}

	instances := make([]types.Instance, 0, len(guests))
	for _, guest := range guests {
		if options.All {
			instances = append(instances, guestToInstance(guest))
		} else {
			for _, tag := range guest.Tags {
				for _, label := range labels {
					if string(tag) == label {
						instances = append(instances, guestToInstance(guest))
						break
					}
				}
			}
		}
	}

	return instances, nil
}

func guestToInstance(guest proxmox.GuestResource) types.Instance {
	var group string = discovery.LabelGroupDefaultValue

	return types.Instance{
		Name:   string(guest.Id),
		Kind:   "guest",
		Status: guest.Status,
		// Replicas: c.Replicas,
		// DesiredReplicas: 1,
		ScalingReplicas: 1,
		Group:           group,
	}
}

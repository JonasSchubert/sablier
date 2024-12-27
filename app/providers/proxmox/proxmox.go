package proxmox

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/Telmate/proxmox-api-go/proxmox"

	"github.com/sablierapp/sablier/app/discovery"
	"github.com/sablierapp/sablier/app/instance"
	"github.com/sablierapp/sablier/app/providers"
	"github.com/sablierapp/sablier/app/types"

	log "github.com/sirupsen/logrus"
)

// Interface guard
var _ providers.Provider = (*ProxmoxProvider)(nil)

type ProxmoxProvider struct {
	Client          proxmox.Client
	desiredReplicas int32
}

func NewProxmoxProvider() (*ProxmoxProvider, error) {
	client, err := cli.Client(cli.Context(), "", "", "", "", "") // TODO
	if err != nil {
		return nil, fmt.Errorf("cannot create proxmox client: %v", err)
	}

	version, err := client.GetVersion(cli.Context())
	if err != nil {
		return nil, fmt.Errorf("cannot connect to proxmox host: %v", err)
	}

	log.Tracef("connection established with proxmox %s", version)

	return &ProxmoxProvider{
		Client:          *client,
		desiredReplicas: 1,
	}, nil
}

func (provider *ProxmoxProvider) GetGroups(ctx context.Context) (map[string][]string, error) {
	guests, err := proxmox.ListGuests(cli.Context(), &provider.Client)
	if err != nil {
		return nil, err
	}

	groups := make(map[string][]string)
	for _, guest := range guests {
		groupName := ""
		index := slices.IndexFunc(guest.Tags, func(tag proxmox.Tag) bool {
			return strings.HasPrefix(string(tag), discovery.LabelEnable)
		})
		if index >= 0 {
			groupName = strings.Replace(string(guest.Tags[index]), discovery.LabelEnable+"=", "", -1)
		}
		if len(groupName) == 0 {
			groupName = discovery.LabelGroupDefaultValue
		}
		group := groups[groupName]
		group = append(group, strings.TrimPrefix(string(guest.Id), "/"))
		groups[groupName] = group
	}

	log.Debug(fmt.Sprintf("%v", groups))

	return groups, nil
}

func (provider *ProxmoxProvider) Start(ctx context.Context, guestId string) error {
	var args = []string{guestId}
	vmRef := proxmox.NewVmRef(cli.ValidateIntIDset(args, "GuestID"))
	_, error := provider.Client.StartVm(cli.Context(), vmRef)

	return error
}

func (provider *ProxmoxProvider) Stop(ctx context.Context, guestId string) error {
	var args = []string{guestId}
	vmRef := proxmox.NewVmRef(cli.ValidateIntIDset(args, "GuestID"))
	_, error := provider.Client.StopVm(cli.Context(), vmRef)

	return error
}

func (provider *ProxmoxProvider) GetState(ctx context.Context, guestId string) (instance.State, error) {
	var args = []string{guestId}
	vmRef := proxmox.NewVmRef(cli.ValidateIntIDset(args, "GuestID"))
	vmState, err := provider.Client.GetVmState(cli.Context(), vmRef)
	if err != nil {
		return instance.State{}, err
	}

	// https://pve.proxmox.com/pve-docs/api-viewer/index.html#/nodes/{node}/lxc/{vmid}/status/current
	// "stopped" or "running"
	var status = vmState["status"].(string)
	switch status {
	case "stopped":
		return instance.NotReadyInstanceState(guestId, 0, provider.desiredReplicas), nil
	case "running":
		return instance.ReadyInstanceState(guestId, provider.desiredReplicas), nil
	default:
		return instance.UnrecoverableInstanceState(guestId, fmt.Sprintf("guest status \"%s\" not handled", status), provider.desiredReplicas), nil
	}
}

func (provider *ProxmoxProvider) NotifyInstanceStopped(ctx context.Context, instance chan<- string) {
	guests, _ := provider.InstanceList(ctx, providers.InstanceListOptions{Labels: []string{}})

	for {
		latestGuests, _ := provider.InstanceList(ctx, providers.InstanceListOptions{Labels: []string{}})
		for _, guest := range guests {
			index := slices.IndexFunc(latestGuests, func(instance types.Instance) bool {
				return instance.Name == guest.Name
			})
			latestGuest := types.Instance{}
			if index >= 0 {
				latestGuest = latestGuests[index]
			}
			if len(latestGuest.Name) == 0 {
				continue
			}
			if guest.Status != latestGuest.Status {
				if latestGuest.Status == "stopped" {
					instance <- strings.TrimPrefix(latestGuest.Name, "/")
				}
			}
		}
		guests = latestGuests
	}
}

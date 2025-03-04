package kubernetes

import (
	"context"
	"github.com/sablierapp/sablier/app/types"
	"github.com/sablierapp/sablier/pkg/provider"
)

func (p *KubernetesProvider) InstanceList(ctx context.Context, options provider.InstanceListOptions) ([]types.Instance, error) {
	deployments, err := p.DeploymentList(ctx)
	if err != nil {
		return nil, err
	}

	statefulSets, err := p.StatefulSetList(ctx)
	if err != nil {
		return nil, err
	}

	return append(deployments, statefulSets...), nil
}

func (p *KubernetesProvider) InstanceGroups(ctx context.Context) (map[string][]string, error) {
	deployments, err := p.DeploymentGroups(ctx)
	if err != nil {
		return nil, err
	}

	statefulSets, err := p.StatefulSetGroups(ctx)
	if err != nil {
		return nil, err
	}

	groups := make(map[string][]string)
	for group, instances := range deployments {
		groups[group] = instances
	}

	for group, instances := range statefulSets {
		groups[group] = append(groups[group], instances...)
	}

	return groups, nil
}

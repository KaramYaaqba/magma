// Copyright (c) 2004-present Facebook All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package resolver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/AlekSi/pointer"
	"github.com/facebookincubator/symphony/graph/graphql/models"
	"github.com/facebookincubator/symphony/graph/resolverutil"
	"github.com/facebookincubator/symphony/pkg/actions"
	"github.com/facebookincubator/symphony/pkg/actions/core"
	"github.com/facebookincubator/symphony/pkg/ent"
	"github.com/facebookincubator/symphony/pkg/ent/equipment"
	"github.com/facebookincubator/symphony/pkg/ent/location"
	"github.com/facebookincubator/symphony/pkg/ent/locationtype"
	"github.com/facebookincubator/symphony/pkg/ent/propertytype"
	"github.com/facebookincubator/symphony/pkg/ent/reportfilter"
	"github.com/facebookincubator/symphony/pkg/ent/service"
	"github.com/facebookincubator/symphony/pkg/ent/servicetype"
	"github.com/facebookincubator/symphony/pkg/viewer"
	"go.uber.org/zap"
)

type queryResolver struct{ resolver }

func (queryResolver) Me(ctx context.Context) (viewer.Viewer, error) {
	return viewer.FromContext(ctx), nil
}

func (r queryResolver) Node(ctx context.Context, id int) (ent.Noder, error) {
	n, err := r.ClientFrom(ctx).Noder(ctx, id)
	if err == nil {
		return n, nil
	}
	r.logger.For(ctx).
		Debug("cannot query node",
			zap.Int("id", id),
			zap.Error(err),
		)
	return nil, ent.MaskNotFound(err)
}

func (r queryResolver) LocationTypes(
	ctx context.Context,
	after *ent.Cursor, first *int,
	before *ent.Cursor, last *int,
) (*ent.LocationTypeConnection, error) {
	return r.ClientFrom(ctx).LocationType.Query().
		Paginate(ctx, after, first, before, last)
}

func (r queryResolver) Locations(
	ctx context.Context, onlyTopLevel *bool,
	types []int, name *string, needsSiteSurvey *bool,
	after *ent.Cursor, first *int,
	before *ent.Cursor, last *int,
	filters []*models.LocationFilterInput,
) (*ent.LocationConnection, error) {
	query := r.ClientFrom(ctx).Location.Query()
	query, err := resolverutil.LocationFilter(query, filters)
	if err != nil {
		return nil, err
	}
	if pointer.GetBool(onlyTopLevel) {
		query = query.Where(location.Not(location.HasParent()))
	}
	if name != nil {
		query = query.Where(location.NameContainsFold(*name))
	}
	if len(types) > 0 {
		query = query.Where(location.HasTypeWith(locationtype.IDIn(types...)))
	}
	if needsSiteSurvey != nil {
		query = query.Where(location.SiteSurveyNeeded(*needsSiteSurvey))
	}
	return query.Paginate(ctx, after, first, before, last)
}

func (r queryResolver) NearestSites(ctx context.Context, latitude, longitude float64, first int) ([]*ent.Location, error) {
	sites := r.ClientFrom(ctx).Location.Query().Where(location.HasTypeWith(locationtype.Site(true))).AllX(ctx)
	var lr locationResolver
	sort.Slice(sites, func(i, j int) bool {
		d1, _ := lr.DistanceKm(ctx, sites[i], latitude, longitude)
		d2, _ := lr.DistanceKm(ctx, sites[j], latitude, longitude)
		return d1 < d2
	})
	if len(sites) < first {
		return sites, nil
	}
	return sites[:first], nil
}

func (r queryResolver) EquipmentTypes(
	ctx context.Context,
	after *ent.Cursor, first *int,
	before *ent.Cursor, last *int,
) (*ent.EquipmentTypeConnection, error) {
	return r.ClientFrom(ctx).EquipmentType.Query().
		Paginate(ctx, after, first, before, last)
}

func (r queryResolver) EquipmentPortTypes(
	ctx context.Context,
	after *ent.Cursor, first *int,
	before *ent.Cursor, last *int,
) (*ent.EquipmentPortTypeConnection, error) {
	return r.ClientFrom(ctx).EquipmentPortType.Query().
		Paginate(ctx, after, first, before, last)
}

func (r queryResolver) EquipmentPortDefinitions(
	ctx context.Context,
	after *ent.Cursor, first *int,
	before *ent.Cursor, last *int,
) (*ent.EquipmentPortDefinitionConnection, error) {
	return r.ClientFrom(ctx).EquipmentPortDefinition.Query().
		Paginate(ctx, after, first, before, last)
}

func (r queryResolver) EquipmentPorts(
	ctx context.Context,
	after *ent.Cursor, first *int,
	before *ent.Cursor, last *int,
	filters []*models.PortFilterInput,
) (*ent.EquipmentPortConnection, error) {
	query := r.ClientFrom(ctx).EquipmentPort.Query()
	query, err := resolverutil.PortFilter(query, filters)
	if err != nil {
		return nil, err
	}
	return query.Paginate(ctx, after, first, before, last)
}

func (r queryResolver) Equipments(
	ctx context.Context,
	after *ent.Cursor, first *int,
	before *ent.Cursor, last *int,
	filters []*models.EquipmentFilterInput,
) (*ent.EquipmentConnection, error) {
	query := r.ClientFrom(ctx).Equipment.Query()
	query, err := resolverutil.EquipmentFilter(query, filters)
	if err != nil {
		return nil, err
	}
	return query.Paginate(ctx, after, first, before, last)
}

func (r queryResolver) WorkOrders(
	ctx context.Context,
	after *ent.Cursor, first *int,
	before *ent.Cursor, last *int,
	orderBy *models.WorkOrderOrder,
	filterBy []*models.WorkOrderFilterInput,
) (*ent.WorkOrderConnection, error) {
	query := r.ClientFrom(ctx).WorkOrder.Query()
	query, err := resolverutil.WorkOrderFilter(query, filterBy)
	if err != nil {
		return nil, err
	}
	return query.Paginate(ctx, after, first, before, last)
}

func (r queryResolver) WorkOrderTypes(
	ctx context.Context,
	after *ent.Cursor, first *int,
	before *ent.Cursor, last *int,
) (*ent.WorkOrderTypeConnection, error) {
	return r.ClientFrom(ctx).WorkOrderType.Query().
		Paginate(ctx, after, first, before, last)
}

func (r queryResolver) Links(
	ctx context.Context,
	after *ent.Cursor, first *int,
	before *ent.Cursor, last *int,
	filters []*models.LinkFilterInput,
) (*ent.LinkConnection, error) {
	query := r.ClientFrom(ctx).Link.Query()
	query, err := resolverutil.LinkFilter(query, filters)
	if err != nil {
		return nil, err
	}
	return query.Paginate(ctx, after, first, before, last)
}

func (r queryResolver) Projects(
	ctx context.Context,
	after *ent.Cursor, first *int,
	before *ent.Cursor, last *int,
	_ []*models.ProjectFilterInput,
) (*ent.ProjectConnection, error) {
	query := r.ClientFrom(ctx).Project.Query()
	return query.Paginate(ctx, after, first, before, last)
}

func (r queryResolver) Services(
	ctx context.Context,
	after *ent.Cursor, first *int,
	before *ent.Cursor, last *int,
	filters []*models.ServiceFilterInput) (*ent.ServiceConnection, error) {
	query := r.ClientFrom(ctx).Service.Query().Where(service.HasTypeWith(servicetype.IsDeleted(false)))
	query, err := resolverutil.ServiceFilter(query, filters)
	if err != nil {
		return nil, err
	}
	return query.Paginate(ctx, after, first, before, last)
}

func (r queryResolver) Users(
	ctx context.Context,
	after *ent.Cursor, first *int,
	before *ent.Cursor, last *int,
	filters []*models.UserFilterInput) (*ent.UserConnection, error) {
	query := r.ClientFrom(ctx).User.Query()
	query, err := resolverutil.UserFilter(query, filters)
	if err != nil {
		return nil, err
	}
	return query.Paginate(ctx, after, first, before, last)
}

func (r queryResolver) UsersGroups(
	ctx context.Context,
	after *ent.Cursor, first *int,
	before *ent.Cursor, last *int,
	filters []*models.UsersGroupFilterInput) (*ent.UsersGroupConnection, error) {
	query := r.ClientFrom(ctx).UsersGroup.Query()
	query, err := resolverutil.UsersGroupFilter(query, filters)
	if err != nil {
		return nil, err
	}
	return query.Paginate(ctx, after, first, before, last)
}

func (r queryResolver) PermissionsPolicies(
	ctx context.Context,
	after *ent.Cursor, first *int,
	before *ent.Cursor, last *int,
	filters []*models.PermissionsPolicyFilterInput) (*ent.PermissionsPolicyConnection, error) {
	query := r.ClientFrom(ctx).PermissionsPolicy.Query()
	query, err := resolverutil.PermissionsPolicyFilter(query, filters)
	if err != nil {
		return nil, err
	}
	return query.Paginate(ctx, after, first, before, last)
}

func (r queryResolver) SearchForNode(
	ctx context.Context, name string,
	_ *ent.Cursor, limit *int,
	_ *ent.Cursor, _ *int,
) (*models.SearchNodesConnection, error) {
	if limit == nil {
		return nil, errors.New("first is a mandatory param")
	}
	client := r.ClientFrom(ctx)
	locations, err := client.Location.Query().
		Where(
			location.Or(
				location.NameContainsFold(name),
				location.ExternalIDContainsFold(name),
			),
		).
		Limit(*limit).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("querying locations: %w", err)
	}

	edges := make([]*models.SearchNodeEdge, len(locations))
	for i, l := range locations {
		edges[i] = &models.SearchNodeEdge{
			Node: l,
		}
	}
	if len(locations) == *limit {
		return &models.SearchNodesConnection{Edges: edges}, nil
	}

	equipments, err := client.Equipment.Query().
		Where(equipment.Or(
			equipment.NameContainsFold(name),
			equipment.ExternalIDContainsFold(name),
		)).
		Limit(*limit - len(locations)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("querying equipment: %w", err)
	}
	for _, e := range equipments {
		edges = append(edges, &models.SearchNodeEdge{
			Node: e,
		})
	}
	return &models.SearchNodesConnection{Edges: edges}, nil
}

func (r queryResolver) PossibleProperties(ctx context.Context, entityType models.PropertyEntity) (pts []*ent.PropertyType, err error) {
	client := r.ClientFrom(ctx)
	switch entityType {
	case models.PropertyEntityEquipment:
		pts, err = client.EquipmentType.Query().QueryPropertyTypes().All(ctx)
	case models.PropertyEntityService:
		pts, err = client.ServiceType.Query().QueryPropertyTypes().All(ctx)
	case models.PropertyEntityLink:
		pts, err = client.EquipmentPortType.Query().QueryLinkPropertyTypes().All(ctx)
	case models.PropertyEntityPort:
		pts, err = client.EquipmentPortType.Query().QueryPropertyTypes().All(ctx)
	case models.PropertyEntityLocation:
		pts, err = client.LocationType.Query().QueryPropertyTypes().All(ctx)
	default:
		return nil, fmt.Errorf("unsupported entity type: %s", entityType)
	}
	if err != nil {
		return nil, fmt.Errorf("querying property types: %w", err)
	}

	type key struct {
		name string
		typ  propertytype.Type
	}
	var (
		groups = map[key]struct{}{}
		types  []*ent.PropertyType
	)
	for _, pt := range pts {
		k := key{pt.Name, pt.Type}
		if _, ok := groups[k]; !ok {
			groups[k] = struct{}{}
			types = append(types, pt)
		}
	}
	return types, nil
}

func (r queryResolver) Surveys(ctx context.Context) ([]*ent.Survey, error) {
	surveys, err := r.ClientFrom(ctx).Survey.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("querying all surveys: %w", err)
	}
	return surveys, nil
}

func (r queryResolver) ServiceTypes(
	ctx context.Context,
	after *ent.Cursor, first *int,
	before *ent.Cursor, last *int,
) (*ent.ServiceTypeConnection, error) {
	return r.ClientFrom(ctx).ServiceType.Query().
		Paginate(ctx, after, first, before, last)
}

func (r queryResolver) Customers(
	ctx context.Context,
	after *ent.Cursor, first *int,
	before *ent.Cursor, last *int,
) (*ent.CustomerConnection, error) {
	return r.ClientFrom(ctx).Customer.Query().
		Paginate(ctx, after, first, before, last)
}

func (r queryResolver) ActionsRules(
	ctx context.Context,
) (*models.ActionsRulesSearchResult, error) {
	results, err := r.ClientFrom(ctx).ActionsRule.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("querying action rules: %w", err)
	}
	return &models.ActionsRulesSearchResult{
		Results: results,
		Count:   len(results),
	}, nil
}

func (r queryResolver) ActionsTriggers(
	ctx context.Context,
) (*models.ActionsTriggersSearchResult, error) {
	triggers := actions.FromContext(ctx).Triggers()
	ret := make([]*models.ActionsTrigger, len(triggers))
	for i, trigger := range triggers {
		ret[i] = &models.ActionsTrigger{
			TriggerID:   trigger.ID(),
			Description: trigger.Description(),
		}
	}
	return &models.ActionsTriggersSearchResult{
		Results: ret,
		Count:   len(ret),
	}, nil
}

func (r queryResolver) ActionsTrigger(
	ctx context.Context, triggerID core.TriggerID,
) (*models.ActionsTrigger, error) {
	trigger, err := actions.FromContext(ctx).
		TriggerForID(triggerID)
	if err != nil {
		return nil, fmt.Errorf("getting trigger: %w", err)
	}
	return &models.ActionsTrigger{
		TriggerID:   triggerID,
		Description: trigger.Description(),
	}, nil
}

func (r queryResolver) ReportFilters(ctx context.Context, entity models.FilterEntity) ([]*ent.ReportFilter, error) {
	rfs, err := r.ClientFrom(ctx).ReportFilter.Query().Where(reportfilter.EntityEQ(reportfilter.Entity(entity))).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("querying report filters for entity %v: %w", entity, err)
	}
	return rfs, nil
}

func (r queryResolver) LatestPythonPackage(ctx context.Context) (*models.LatestPythonPackageResult, error) {
	packages, err := r.PythonPackages(ctx)
	if err != nil {
		return nil, err
	}
	if len(packages) == 0 {
		return nil, nil
	}
	lastBreakingChange := len(packages) - 1
	for i, pkg := range packages {
		if pkg.HasBreakingChange {
			lastBreakingChange = i
			break
		}
	}
	return &models.LatestPythonPackageResult{
		LastPythonPackage:         packages[0],
		LastBreakingPythonPackage: packages[lastBreakingChange],
	}, nil
}

func (queryResolver) PythonPackages(context.Context) ([]*models.PythonPackage, error) {
	var (
		packages []models.PythonPackage
		res      []*models.PythonPackage
	)
	if err := json.Unmarshal([]byte(PyinventoryConsts), &packages); err != nil {
		return nil, fmt.Errorf("decoding python packages: %w", err)
	}
	for _, p := range packages {
		p := p
		res = append(res, &p)
	}
	return res, nil
}

func (r queryResolver) Vertex(ctx context.Context, id int) (*ent.Node, error) {
	return r.ClientFrom(ctx).Node(ctx, id)
}

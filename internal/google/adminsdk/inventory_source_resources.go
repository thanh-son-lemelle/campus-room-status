package adminsdk

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"campus-room-status/internal/domain"
	admin "google.golang.org/api/admin/directory/v1"
)

func (s *InventorySource) listCalendarResources(ctx context.Context) ([]domain.Room, error) {
	pageToken := ""
	visitedTokens := make(map[string]struct{})
	roomsByCode := make(map[string]domain.Room)

	for {
		call := s.service.Resources.Calendars.List(s.customer).MaxResults(s.pageSize)
		if pageToken != "" {
			if _, exists := visitedTokens[pageToken]; exists {
				return nil, fmt.Errorf("detected repeated page token %q on calendars endpoint", pageToken)
			}
			visitedTokens[pageToken] = struct{}{}
			call = call.PageToken(pageToken)
		}

		response, err := call.Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("calendars list request failed: %w", err)
		}
		if response == nil {
			break
		}

		for _, resource := range response.Items {
			mapped, fields, ok := mapCalendarResource(resource)
			if !ok {
				continue
			}
			s.recordObservedFields(fields)
			roomsByCode[mapped.Code] = mapped
		}

		if response.NextPageToken == "" {
			break
		}
		pageToken = response.NextPageToken
	}

	rooms := make([]domain.Room, 0, len(roomsByCode))
	for _, room := range roomsByCode {
		rooms = append(rooms, room)
	}
	sort.SliceStable(rooms, func(i, j int) bool {
		return rooms[i].Code < rooms[j].Code
	})

	return rooms, nil
}

func (s *InventorySource) recordObservedFields(fields []string) {
	s.observedFieldsMu.Lock()
	defer s.observedFieldsMu.Unlock()

	for _, field := range fields {
		s.observedFields[field] = struct{}{}
	}
}

func mapCalendarResource(resource *admin.CalendarResource) (domain.Room, []string, bool) {
	if resource == nil {
		return domain.Room{}, nil, false
	}

	resourceName := strings.TrimSpace(resource.ResourceName)
	resourceEmail := strings.TrimSpace(resource.ResourceEmail)
	generatedResourceName := strings.TrimSpace(resource.GeneratedResourceName)

	code := firstNonEmpty(
		generatedResourceName,
		resourceName,
		codeFromEmail(resourceEmail),
	)
	if code == "" {
		return domain.Room{}, nil, false
	}

	floor, _ := strconv.Atoi(strings.TrimSpace(resource.FloorName))
	resourceType := firstNonEmpty(resource.ResourceType, resource.ResourceCategory, "room")

	return domain.Room{
		Code:          code,
		ResourceEmail: resourceEmail,
		Name:          firstNonEmpty(resourceName, generatedResourceName, code),
		Building:      strings.TrimSpace(resource.BuildingId),
		Floor:         floor,
		Capacity:      int(resource.Capacity),
		Type:          resourceType,
		Status:        "available",
	}, observedFieldsFromResource(resource), true
}

func observedFieldsFromResource(resource *admin.CalendarResource) []string {
	fields := make([]string, 0, 16)

	addField := func(name string, shouldAdd bool) {
		if shouldAdd {
			fields = append(fields, name)
		}
	}

	addField("generatedResourceName", strings.TrimSpace(resource.GeneratedResourceName) != "")
	addField("resourceName", strings.TrimSpace(resource.ResourceName) != "")
	addField("resourceEmail", strings.TrimSpace(resource.ResourceEmail) != "")
	addField("capacity", resource.Capacity > 0)
	addField("resourceType", strings.TrimSpace(resource.ResourceType) != "")
	addField("resourceCategory", strings.TrimSpace(resource.ResourceCategory) != "")
	addField("buildingId", strings.TrimSpace(resource.BuildingId) != "")
	addField("floorName", strings.TrimSpace(resource.FloorName) != "")
	addField("floorSection", strings.TrimSpace(resource.FloorSection) != "")
	addField("resourceDescription", strings.TrimSpace(resource.ResourceDescription) != "")
	addField("userVisibleDescription", strings.TrimSpace(resource.UserVisibleDescription) != "")
	addField("featureInstances", hasAnyValue(resource.FeatureInstances))

	sort.Strings(fields)
	return fields
}

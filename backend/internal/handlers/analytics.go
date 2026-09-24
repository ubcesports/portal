package handlers

import (
	"encoding/csv"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/ubcesports/memberships/internal/dto"
	"github.com/ubcesports/memberships/internal/service"
	"github.com/ubcesports/memberships/internal/util"
)

type AnalyticsHandler struct {
	analyticsService *service.AnalyticsService
}

func NewAnalyticsHandler(analyticsService *service.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{analyticsService: analyticsService}
}

/*
Returns current/all-time summary stats for the given filters.

API URL: GET /admin/analytics/summary

Args (query params):

	program_name: optional membership program name
	tier_id: optional repeated membership tier UUIDs
	is_student: optional boolean student status
	purchase_type: optional purchase type (new or upgrade)

Returns:

	dto.AnalyticsSummaryDTO (HTTP 200)

Raises:

	400: invalid filter value
	401: user is not authenticated
	403: user is not an admin
	500: summary could not be loaded
*/
func (h *AnalyticsHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())

	filters, err := parseAnalyticsFilters(r, false)
	if err != nil {
		util.WriteApiResponse(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), requestID)
		return
	}

	summary, err := h.analyticsService.GetSummary(r.Context(), filters)
	if err != nil {
		slog.ErrorContext(r.Context(), "unable to load analytics summary", "error", err, "request_id", requestID)
		util.WriteApiResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to load analytics summary.", requestID)
		return
	}

	util.WriteJson(w, http.StatusOK, summary)
}

/*
Returns the number of completed membership purchases per period.

API URL: GET /admin/analytics/memberships-over-time

Args (query params):

	program_name, tier_id, is_student: same as GetSummary
	granularity: week, month, or year (default week)
	periods: number of trailing periods to include (default 12), ignored when since_inception is set
	since_inception: "true" to return the full history instead of the last N periods

Returns:

	[]dto.MembershipsBoughtPointDTO (HTTP 200)

Raises:

	400: invalid filter value
	401: user is not authenticated
	403: user is not an admin
	500: data could not be loaded
*/
func (h *AnalyticsHandler) GetMembershipsBoughtOverTime(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())

	filters, err := parseAnalyticsFilters(r, true)
	if err != nil {
		util.WriteApiResponse(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), requestID)
		return
	}

	points, err := h.analyticsService.GetMembershipsBoughtOverTime(r.Context(), filters)
	if err != nil {
		slog.ErrorContext(r.Context(), "unable to load memberships bought over time", "error", err, "request_id", requestID)
		util.WriteApiResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to load memberships bought over time.", requestID)
		return
	}

	util.WriteJson(w, http.StatusOK, points)
}

/*
Returns revenue collected per period.

API URL: GET /admin/analytics/revenue-over-time

Args (query params):

	program_name, tier_id, is_student, purchase_type: same as GetSummary
	granularity, periods, since_inception: same as GetMembershipsBoughtOverTime

Returns:

	[]dto.RevenuePointDTO (HTTP 200)

Raises:

	400: invalid filter value
	401: user is not authenticated
	403: user is not an admin
	500: data could not be loaded
*/
func (h *AnalyticsHandler) GetRevenueOverTime(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())

	filters, err := parseAnalyticsFilters(r, true)
	if err != nil {
		util.WriteApiResponse(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), requestID)
		return
	}

	points, err := h.analyticsService.GetRevenueOverTime(r.Context(), filters)
	if err != nil {
		slog.ErrorContext(r.Context(), "unable to load revenue over time", "error", err, "request_id", requestID)
		util.WriteApiResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to load revenue over time.", requestID)
		return
	}

	util.WriteJson(w, http.StatusOK, points)
}

/*
Exports the memberships-bought and revenue time series for the given filters
and date range as a single CSV, one row per period.

API URL: GET /admin/analytics/export

Args (query params):

	Same as GetRevenueOverTime.

Returns:

	membership-analytics.csv: CSV file (HTTP 200)

Raises:

	400: invalid filter value
	401: user is not authenticated
	403: user is not an admin
	500: data could not be exported
*/
func (h *AnalyticsHandler) ExportCSV(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())

	filters, err := parseAnalyticsFilters(r, true)
	if err != nil {
		util.WriteApiResponse(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), requestID)
		return
	}

	rows, err := h.analyticsService.ExportCSV(r.Context(), filters)
	if err != nil {
		slog.ErrorContext(r.Context(), "unable to export membership analytics", "error", err, "request_id", requestID)
		util.WriteApiResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to export membership analytics.", requestID)
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="membership-analytics.csv"`)

	writer := csv.NewWriter(w)
	if err := writer.Write([]string{
		"Period Start",
		"Period End",
		"Memberships Bought",
		"Revenue (CAD)",
	}); err != nil {
		slog.ErrorContext(r.Context(), "unable to write CSV header", "error", err, "request_id", requestID)
		return
	}

	for _, row := range rows {
		if err := writer.Write([]string{
			row.PeriodStart.Format(time.RFC3339),
			row.PeriodEnd.Format(time.RFC3339),
			strconv.FormatInt(row.MembershipsBought, 10),
			fmt.Sprintf("%.2f", float64(row.RevenueCents)/100),
		}); err != nil {
			slog.ErrorContext(r.Context(), "unable to write CSV row", "error", err, "request_id", requestID)
			return
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		slog.ErrorContext(r.Context(), "unable to flush CSV response", "error", err, "request_id", requestID)
		return
	}
}

/*
	Private functions
*/

func parseAnalyticsFilters(r *http.Request, requireGranularity bool) (service.AnalyticsFilters, error) {
	query := r.URL.Query()
	filters := service.AnalyticsFilters{
		TierIDs: query["tier_id"],
		Periods: 12,
	}

	if programName := query.Get("program_name"); programName != "" {
		filters.ProgramName = &programName
	}

	for _, tierID := range filters.TierIDs {
		if _, err := util.GetValidatedUUID(tierID); err != nil {
			return service.AnalyticsFilters{}, errors.New("invalid membership tier ID")
		}
	}

	if value := query.Get("is_student"); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return service.AnalyticsFilters{}, errors.New("is_student must be true or false")
		}
		filters.IsStudent = &parsed
	}

	if value := query.Get("purchase_type"); value != "" {
		switch dto.PurchaseType(value) {
		case dto.PurchaseNew, dto.PurchaseUpgrade:
			filters.PurchaseType = &value
		default:
			return service.AnalyticsFilters{}, errors.New("purchase_type must be new or upgrade")
		}
	}

	if !requireGranularity {
		return filters, nil
	}

	switch granularity := query.Get("granularity"); granularity {
	case "", "week":
		filters.Granularity = "week"
	case "month", "year":
		filters.Granularity = granularity
	default:
		return service.AnalyticsFilters{}, errors.New("granularity must be week, month, or year")
	}

	if query.Get("since_inception") == "true" {
		filters.SinceInception = true
		return filters, nil
	}

	if value := query.Get("periods"); value != "" {
		parsed, err := strconv.ParseInt(value, 10, 32)
		if err != nil || parsed <= 0 {
			return service.AnalyticsFilters{}, errors.New("periods must be a positive integer")
		}
		filters.Periods = int(parsed)
	}

	return filters, nil
}

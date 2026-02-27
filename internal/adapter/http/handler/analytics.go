package handler

import (
	"net/http"
	"task_vault/internal/app/query"

	"github.com/go-chi/chi/v5"
)

type AnalyticsHandler struct {
	analytics *query.TeamAnalyticsHandler
}

func NewAnalyticsHandler(analytics *query.TeamAnalyticsHandler) *AnalyticsHandler {
	return &AnalyticsHandler{analytics: analytics}
}

// @Summary      Статистика команд
// @Description  Для каждой команды: название, количество участников, количество задач done за 7 дней
// @Tags         analytics
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  swagTeamStatsResponse
// @Failure      500  {object}  swagErrorResponse
// @Router       /analytics/teams [get]
func (h *AnalyticsHandler) TeamStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.analytics.TeamStats(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ошибка получения статистики")
		return
	}

	result := make([]TeamStatResponse, 0, len(stats))
	for _, s := range stats {
		result = append(result, TeamStatResponse{
			TeamID:       s.TeamID,
			TeamName:     s.TeamName,
			MembersCount: s.MembersCount,
			DoneLastWeek: s.DoneLastWeek,
		})
	}
	writeJSON(w, http.StatusOK, result)
}

// @Summary      Топ контрибьюторы
// @Description  Топ-3 пользователя по количеству созданных задач в команде за месяц
// @Tags         analytics
// @Produce      json
// @Security     BearerAuth
// @Param        id  path      string true "ID команды"
// @Success      200 {object}  swagUserRankResponse
// @Failure      400 {object}  swagErrorResponse
// @Failure      500 {object}  swagErrorResponse
// @Router       /analytics/teams/{id}/top [get]
func (h *AnalyticsHandler) TopContributors(w http.ResponseWriter, r *http.Request) {
	teamID := chi.URLParam(r, "id")
	if teamID == "" {
		writeError(w, http.StatusBadRequest, "невалидный ID команды")
		return
	}

	ranks, err := h.analytics.TopContributors(r.Context(), teamID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ошибка получения контрибьюторов")
		return
	}

	result := make([]UserRankResponse, 0, len(ranks))
	for _, r := range ranks {
		result = append(result, UserRankResponse{
			UserID:       r.UserID,
			UserName:     r.UserName,
			TeamID:       r.TeamID,
			TasksCreated: r.TasksCreated,
			Rank:         r.Rank,
		})
	}
	writeJSON(w, http.StatusOK, result)
}

// @Summary      Задачи с невалидным исполнителем
// @Description  Задачи, где assignee не является членом команды задачи
// @Tags         analytics
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object}  swagTaskListOnlyResponse
// @Failure      500 {object}  swagErrorResponse
// @Router       /analytics/orphaned-assignees [get]
func (h *AnalyticsHandler) OrphanedAssignees(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.analytics.OrphanedAssignees(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ошибка валидации целостности")
		return
	}

	result := make([]TaskResponse, 0, len(tasks))
	for _, t := range tasks {
		result = append(result, toTaskResponse(&t))
	}
	writeJSON(w, http.StatusOK, result)
}

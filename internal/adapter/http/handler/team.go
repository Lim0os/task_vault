package handler

import (
	"errors"
	"net/http"
	"task_vault/internal/adapter/http/middleware"
	"task_vault/internal/app/command"
	"task_vault/internal/app/query"
	"task_vault/internal/domain"

	"github.com/go-chi/chi/v5"
)

type TeamHandler struct {
	createTeam   *command.CreateTeamHandler
	inviteMember *command.InviteMemberHandler
	sendInvite   *command.SendInviteHandler
	getTeams     *query.GetTeamsHandler
}

func NewTeamHandler(
	createTeam *command.CreateTeamHandler,
	inviteMember *command.InviteMemberHandler,
	sendInvite *command.SendInviteHandler,
	getTeams *query.GetTeamsHandler,
) *TeamHandler {
	return &TeamHandler{
		createTeam:   createTeam,
		inviteMember: inviteMember,
		sendInvite:   sendInvite,
		getTeams:     getTeams,
	}
}

// @Summary      Создание команды
// @Description  Создает новую команду. Текущий пользователь становится владельцем.
// @Tags         teams
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        input body      CreateTeamRequest true "Название команды"
// @Success      201   {object}  swagTeamResponse
// @Failure      400   {object}  swagErrorResponse
// @Failure      500   {object}  swagErrorResponse
// @Router       /teams [post]
func (h *TeamHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateTeamRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "невалидный JSON")
		return
	}
	if verr := req.Validate(); verr != nil {
		writeValidationError(w, verr)
		return
	}

	userID := middleware.GetUserID(r.Context())
	team, err := h.createTeam.Handle(r.Context(), command.CreateTeamInput{
		Name:      req.Name,
		CreatedBy: userID,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ошибка создания команды")
		return
	}

	writeJSON(w, http.StatusCreated, TeamResponse{
		ID:        team.ID,
		Name:      team.Name,
		CreatedBy: team.CreatedBy,
	})
}

// @Summary      Список команд
// @Description  Возвращает команды текущего пользователя
// @Tags         teams
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  swagTeamListResponse
// @Failure      500  {object}  swagErrorResponse
// @Router       /teams [get]
func (h *TeamHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	teams, err := h.getTeams.Handle(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ошибка получения команд")
		return
	}

	result := make([]TeamResponse, 0, len(teams))
	for _, t := range teams {
		result = append(result, TeamResponse{
			ID:        t.ID,
			Name:      t.Name,
			CreatedBy: t.CreatedBy,
		})
	}
	writeJSON(w, http.StatusOK, result)
}

// @Summary      Приглашение в команду
// @Description  Приглашает пользователя в команду по email. Требуется роль owner или admin.
// @Tags         teams
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string        true "ID команды"
// @Param        input body      InviteRequest true "Email приглашаемого"
// @Success      200   {object}  swagStatusResponse
// @Failure      400   {object}  swagErrorResponse
// @Failure      403   {object}  swagErrorResponse
// @Failure      404   {object}  swagErrorResponse
// @Failure      409   {object}  swagErrorResponse
// @Failure      500   {object}  swagErrorResponse
// @Router       /teams/{id}/invite [post]
func (h *TeamHandler) Invite(w http.ResponseWriter, r *http.Request) {
	teamID := chi.URLParam(r, "id")
	if teamID == "" {
		writeError(w, http.StatusBadRequest, "невалидный ID команды")
		return
	}

	var req InviteRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "невалидный JSON")
		return
	}
	if verr := req.Validate(); verr != nil {
		writeValidationError(w, verr)
		return
	}

	userID := middleware.GetUserID(r.Context())
	err := h.inviteMember.Handle(r.Context(), command.InviteMemberInput{
		TeamID:      teamID,
		InvitedByID: userID,
		UserEmail:   req.Email,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNoPermission):
			writeError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, domain.ErrUserNotFound):
			writeError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, domain.ErrAlreadyMember):
			writeError(w, http.StatusConflict, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "ошибка приглашения")
		}
		return
	}

	_ = h.sendInvite.Handle(r.Context(), command.SendInviteInput{
		TeamID:    teamID,
		SenderID:  userID,
		UserEmail: req.Email,
	})

	writeJSON(w, http.StatusOK, map[string]string{"status": "приглашение отправлено"})
}
